package handler

import (
	"context"

	"cinema-booking/internal/lock"
	"cinema-booking/internal/model"
	"cinema-booking/internal/mq"
	"cinema-booking/internal/repository"
	"cinema-booking/internal/ws"
)

type Handler struct {
	Repo     *repository.MongoRepo
	Lock     *lock.Manager
	Hub      *ws.Hub
	Pub      *mq.Publisher
	JWTSecret string
	OnAudit  func(event string, payload map[string]any)
}

func (h *Handler) audit(event string, payload map[string]any) {
	if h.OnAudit != nil {
		h.OnAudit(event, payload)
	}
}

func (h *Handler) seatState(ctx context.Context, screeningID string, bookings []*model.Booking, row, col int) model.Seat {
	st := model.Seat{Row: row, Col: col, Status: model.SeatAvailable}
	for _, b := range bookings {
		if b.SeatRow == row && b.SeatCol == col {
			if b.Status == "CONFIRMED" {
				st.Status = model.SeatBooked
				st.UserID = b.UserID
				return st
			}
			if b.Status == "PENDING" && b.LockID != "" {
				lockID, _ := h.Lock.GetLockID(ctx, screeningID, row, col)
				if lockID == b.LockID {
					st.Status = model.SeatLocked
					st.LockID = b.LockID
					st.UserID = b.UserID
					return st
				}
			}
		}
	}
	return st
}
