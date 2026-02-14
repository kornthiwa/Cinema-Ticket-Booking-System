package handler

import (
	"net/http"
	"time"

	"cinema-booking/internal/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) LockSeat(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	screeningID := c.Param("id")
	var body struct {
		Row int `json:"row" binding:"min=0"`
		Col int `json:"col" binding:"min=0"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s, err := h.Repo.GetScreening(c.Request.Context(), screeningID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "screening not found"})
		return
	}
	if body.Row < 0 || body.Row >= s.Rows || body.Col < 0 || body.Col >= s.Cols {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid seat"})
		return
	}
	lockID, err := h.Lock.Acquire(c.Request.Context(), screeningID, body.Row, body.Col)
	if err != nil {
		h.audit(model.EventLockFailed, map[string]any{"screening_id": screeningID, "row": body.Row, "col": body.Col, "error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "lock failed"})
		return
	}
	if lockID == "" {
		c.JSON(http.StatusConflict, gin.H{"error": "seat already locked or booked"})
		return
	}
	b := &model.Booking{
		ScreeningID: screeningID,
		UserID:      userID,
		SeatRow:     body.Row,
		SeatCol:     body.Col,
		Status:      "PENDING",
		LockID:      lockID,
		CreatedAt:   time.Now(),
	}
	if err := h.Repo.CreateBooking(c.Request.Context(), b); err != nil {
		_ = h.Lock.Release(c.Request.Context(), screeningID, body.Row, body.Col, lockID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Broadcast seat update so other users see LOCKED in real-time
	h.Hub.BroadcastSeatUpdate("screening:"+screeningID, h.seatState(c.Request.Context(), screeningID, []*model.Booking{b}, body.Row, body.Col))
	c.JSON(http.StatusOK, gin.H{"lock_id": lockID, "expires_in_seconds": 300, "booking_id": b.ID.Hex()})
}

func (h *Handler) ConfirmPayment(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var body struct {
		BookingID string `json:"booking_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	b, err := h.Repo.GetBookingByID(c.Request.Context(), body.BookingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}
	if b.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your booking"})
		return
	}
	if b.Status != "PENDING" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "booking already " + b.Status})
		return
	}
	// Verify lock still held
	lockID, _ := h.Lock.GetLockID(c.Request.Context(), b.ScreeningID, b.SeatRow, b.SeatCol)
	if lockID != b.LockID {
		c.JSON(http.StatusConflict, gin.H{"error": "lock expired"})
		return
	}
	if err := h.Repo.ConfirmBooking(c.Request.Context(), body.BookingID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Keep key but we consider seat BOOKED; optionally delete lock or let it expire
	_ = h.Lock.Release(c.Request.Context(), b.ScreeningID, b.SeatRow, b.SeatCol, b.LockID)
	h.audit(model.EventBookingSuccess, map[string]any{"booking_id": body.BookingID, "user_id": userID, "screening_id": b.ScreeningID})
	_ = h.Pub.PublishBookingSuccess(c.Request.Context(), b.ScreeningID, userID, body.BookingID, b.SeatRow, b.SeatCol)
	// Broadcast so seat shows BOOKED
	bookings, _ := h.Repo.ListBookings(c.Request.Context(), map[string]interface{}{"screening_id": b.ScreeningID})
	h.Hub.BroadcastSeatUpdate("screening:"+b.ScreeningID, h.seatState(c.Request.Context(), b.ScreeningID, bookings, b.SeatRow, b.SeatCol))
	c.JSON(http.StatusOK, gin.H{"status": "confirmed"})
}
