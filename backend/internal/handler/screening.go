package handler

import (
	"net/http"
	"time"

	"cinema-booking/internal/model"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
