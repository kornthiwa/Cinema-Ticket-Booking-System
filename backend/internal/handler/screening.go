package handler

import (
	"net/http"
	"time"

	"cinema-booking/internal/model"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SeatLockInfo is returned for a locked seat (who, when, unlocks when).
type SeatLockInfo struct {
	Row       int       `json:"row"`
	Col       int       `json:"col"`
	UserID    string    `json:"user_id"`
	BookingID string    `json:"booking_id,omitempty"`
	LockedAt  time.Time `json:"locked_at"`
	UnlocksAt time.Time `json:"unlocks_at"`
}

// SeatBookedInfo is returned for a confirmed booking.
type SeatBookedInfo struct {
	Row      int        `json:"row"`
	Col      int        `json:"col"`
	UserID   string     `json:"user_id"`
	BookedAt *time.Time `json:"booked_at,omitempty"`
}

func (h *Handler) ListScreenings(c *gin.Context) {
	list, err := h.Repo.ListScreenings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) GetScreening(c *gin.Context) {
	id := c.Param("id")
	s, err := h.Repo.GetScreening(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "screening not found"})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *Handler) GetSeatMap(c *gin.Context) {
	id := c.Param("id")
	s, err := h.Repo.GetScreening(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "screening not found"})
		return
	}
	bookings, _ := h.Repo.ListBookings(c.Request.Context(), map[string]interface{}{"screening_id": id})
	ctx := c.Request.Context()
	seats := make([][]model.Seat, s.Rows)
	for r := 0; r < s.Rows; r++ {
		seats[r] = make([]model.Seat, s.Cols)
		for col := 0; col < s.Cols; col++ {
			seats[r][col] = h.seatState(ctx, id, bookings, r, col)
		}
	}
	c.JSON(http.StatusOK, gin.H{"screening": s, "seats": seats})
}

// GetSeatDetails returns who locked/booked which seats and when (for listing on ScreeningList).
func (h *Handler) GetSeatDetails(c *gin.Context) {
	id := c.Param("id")
	s, err := h.Repo.GetScreening(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "screening not found"})
		return
	}
	bookings, _ := h.Repo.ListBookings(c.Request.Context(), map[string]interface{}{"screening_id": id})
	ttlSec := 300
	if h.LockTTLSeconds > 0 {
		ttlSec = h.LockTTLSeconds
	}
	ttl := time.Duration(ttlSec) * time.Second
	var locked []SeatLockInfo
	var booked []SeatBookedInfo
	ctx := c.Request.Context()
	for _, b := range bookings {
		if b.Status == "CONFIRMED" {
			booked = append(booked, SeatBookedInfo{
				Row: b.SeatRow, Col: b.SeatCol, UserID: b.UserID,
				BookedAt: b.ConfirmedAt,
			})
			if booked[len(booked)-1].BookedAt == nil {
				booked[len(booked)-1].BookedAt = &b.CreatedAt
			}
			continue
		}
		if b.Status == "PENDING" && b.LockID != "" {
			lockID, _ := h.Lock.GetLockID(ctx, id, b.SeatRow, b.SeatCol)
			if lockID == b.LockID {
				unlocksAt := b.CreatedAt.Add(ttl)
				locked = append(locked, SeatLockInfo{
					Row: b.SeatRow, Col: b.SeatCol, UserID: b.UserID,
					BookingID: b.ID.Hex(), LockedAt: b.CreatedAt, UnlocksAt: unlocksAt,
				})
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"screening": s,
		"locked":    locked,
		"booked":    booked,
	})
}

func (h *Handler) CreateScreening(c *gin.Context) {
	var body struct {
		MovieID   string `json:"movie_id" binding:"required"`
		MovieName string `json:"movie_name" binding:"required"`
		ScreenAt  string `json:"screen_at" binding:"required"`
		Rows      int    `json:"rows" binding:"required,min=1"`
		Cols      int    `json:"cols" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, err := time.Parse(time.RFC3339, body.ScreenAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid screen_at"})
		return
	}
	s := &model.Screening{
		ID:        primitive.NewObjectID(),
		MovieID:   body.MovieID,
		MovieName: body.MovieName,
		ScreenAt:  t,
		Rows:      body.Rows,
		Cols:      body.Cols,
		CreatedAt: time.Now(),
	}
	if err := h.Repo.CreateScreening(c.Request.Context(), s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, s)
}
