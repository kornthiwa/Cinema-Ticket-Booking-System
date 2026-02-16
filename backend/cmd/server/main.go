package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"cinema-booking/config"
	"cinema-booking/internal/auth"
	"cinema-booking/internal/handler"
	"cinema-booking/internal/lock"
	"cinema-booking/internal/middleware"
	"cinema-booking/internal/model"
	"cinema-booking/internal/mq"
	"cinema-booking/internal/repository"
	"cinema-booking/internal/seed"
	"cinema-booking/internal/ws"
	"cinema-booking/internal/worker"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatal("mongo connect:", err)
	}
	defer mongoClient.Disconnect(ctx)
	db := mongoClient.Database("cinema")

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("redis ping:", err)
	}
	defer rdb.Close()

	lockMgr := lock.NewManager(rdb, cfg.LockTTLSeconds)
	repo := repository.NewMongoRepo(db)

	seed.Run(ctx, repo)

	hub := ws.NewHub()
	go hub.Run()

	pub := mq.NewPublisher(rdb)
	onAudit := func(event string, payload map[string]any) {
		if err := repo.InsertAuditLog(ctx, event, payload); err != nil {
			log.Printf("audit insert: %v", err)
		}
	}
	sub := mq.NewSubscriber(rdb, func(ev mq.Event) {
		onAudit(ev.Type, ev.Payload)
		if ev.Type == "BOOKING_SUCCESS" {
			log.Printf("[MQ] Booking success notification (mock): %+v", ev.Payload)
			if sid, ok := ev.Payload["screening_id"].(string); ok {
				hub.BroadcastNotification("screening:"+sid, ev.Type, ev.Payload)
			}
		}
		if ev.Type == "SEAT_RELEASED" {
			if sid, ok := ev.Payload["screening_id"].(string); ok {
				hub.BroadcastNotification("screening:"+sid, ev.Type, ev.Payload)
			}
		}
	})
	go sub.Run(ctx)

	go worker.RunLockExpiry(ctx, repo, lockMgr, pub, hub, onAudit)

	h := &handler.Handler{
		Repo:           repo,
		Lock:           lockMgr,
		Hub:            hub,
		Pub:            pub,
		JWTSecret:      cfg.JWTSecret,
		LockTTLSeconds: cfg.LockTTLSeconds,
		OnAudit:        onAudit,
	}

	r := gin.Default()
	r.Use(corsMiddleware())

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Cinema Booking API",
			"docs":    "Frontend (app): http://localhost (port 80). API: /auth/login, /api/screenings, /admin/*",
		})
	})

	r.POST("/auth/login", h.Login)
	r.POST("/auth/register", h.Register)

	api := r.Group("/api")
	api.Use(middleware.Auth(cfg.JWTSecret))
	{
		api.GET("/screenings", h.ListScreenings)
		api.GET("/screenings/:id", h.GetScreening)
		api.GET("/screenings/:id/seats", h.GetSeatMap)
		api.GET("/screenings/:id/seat-details", h.GetSeatDetails)
		api.GET("/screenings/:id/ws", h.ServeWS)
		api.POST("/screenings/:id/lock", h.LockSeat)
		api.POST("/bookings/confirm", h.ConfirmPayment)
	}

	admin := r.Group("/admin")
	admin.Use(middleware.Auth(cfg.JWTSecret), middleware.AdminOnly())
	{
		admin.GET("/bookings", h.ListBookingsAdmin)
		admin.GET("/audit-logs", h.ListAuditLogs)
		admin.GET("/ws", h.ServeAdminWS)
	}

	r.POST("/admin/login", func(c *gin.Context) {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		u, err := repo.GetUserByEmail(c.Request.Context(), body.Email)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		if u.Role != model.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access only"})
			return
		}
		if u.PasswordHash == "" || !auth.CheckPassword(u.PasswordHash, body.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		token, err := auth.IssueToken(cfg.JWTSecret, u.ID, u.Email, "ADMIN")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token, "user_id": u.ID, "role": "ADMIN"})
	})

	admin.POST("/screenings", h.CreateScreening)

	addr := ":" + strconv.Itoa(cfg.ServerPort)
	if err := r.Run(addr); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
