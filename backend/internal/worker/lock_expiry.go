package worker

import (
	"context"
	"log"
	"time"

	"cinema-booking/internal/lock"
	"cinema-booking/internal/model"
	"cinema-booking/internal/mq"
	"cinema-booking/internal/repository"
	"cinema-booking/internal/ws"
	"go.mongodb.org/mongo-driver/bson"
)

const lockTTL = 5 * time.Minute

// RunLockExpiry periodically marks expired PENDING bookings as TIMEOUT, releases Redis lock, audits, publishes and broadcasts.
func RunLockExpiry(ctx context.Context, repo *repository.MongoRepo, lockMgr *lock.Manager, pub *mq.Publisher, hub *ws.Hub, onAudit func(string, map[string]any)) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cutoff := time.Now().Add(-lockTTL)
			list, err := repo.ListBookings(ctx, bson.M{"status": "PENDING", "created_at": bson.M{"$lt": cutoff}})
			if err != nil {
				log.Printf("lock_expiry: list: %v", err)
				continue
			}
			for _, b := range list {
				_ = lockMgr.Release(ctx, b.ScreeningID, b.SeatRow, b.SeatCol, b.LockID)
				// Only set TIMEOUT if still PENDING (avoid overwriting CONFIRMED after race)
				updated, _ := repo.SetBookingStatusIfPending(ctx, b.ID.Hex(), "TIMEOUT")
				if !updated {
					continue
				}
				if onAudit != nil {
					onAudit(model.EventBookingTimeout, map[string]any{"booking_id": b.ID.Hex(), "screening_id": b.ScreeningID, "seat_row": b.SeatRow, "seat_col": b.SeatCol})
				}
				_ = pub.PublishSeatReleased(ctx, b.ScreeningID, b.SeatRow, b.SeatCol)
				bookings, _ := repo.ListBookings(ctx, bson.M{"screening_id": b.ScreeningID})
				seat := seatStateFor(b.ScreeningID, bookings, b.SeatRow, b.SeatCol)
				hub.BroadcastSeatUpdate("screening:"+b.ScreeningID, seat)
				hub.BroadcastAdmin("REFRESH", nil)
			}
		}
	}
}

func seatStateFor(screeningID string, bookings []*model.Booking, row, col int) model.Seat {
	st := model.Seat{Row: row, Col: col, Status: model.SeatAvailable}
	for _, b := range bookings {
		if b.SeatRow == row && b.SeatCol == col && b.Status == "CONFIRMED" {
			st.Status = model.SeatBooked
			st.UserID = b.UserID
			return st
		}
	}
	return st
}
