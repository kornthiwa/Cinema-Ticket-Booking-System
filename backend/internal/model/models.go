package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SeatStatus string

const (
	SeatAvailable SeatStatus = "AVAILABLE"
	SeatLocked    SeatStatus = "LOCKED"
	SeatBooked    SeatStatus = "BOOKED"
)

type UserRole string

const (
	RoleUser  UserRole = "USER"
	RoleAdmin UserRole = "ADMIN"
)

type Screening struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MovieID   string             `bson:"movie_id" json:"movie_id"`
	MovieName string             `bson:"movie_name" json:"movie_name"`
	ScreenAt  time.Time          `bson:"screen_at" json:"screen_at"`
	Rows      int                `bson:"rows" json:"rows"`
	Cols      int                `bson:"cols" json:"cols"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type Seat struct {
	Row    int        `bson:"row" json:"row"`
	Col    int        `bson:"col" json:"col"`
	Status SeatStatus `bson:"status" json:"status"`
	LockID string     `bson:"lock_id,omitempty" json:"lock_id,omitempty"`
	UserID string     `bson:"user_id,omitempty" json:"user_id,omitempty"`
}

type Booking struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ScreeningID string             `bson:"screening_id" json:"screening_id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	SeatRow     int                `bson:"seat_row" json:"seat_row"`
	SeatCol     int                `bson:"seat_col" json:"seat_col"`
	Status      string             `bson:"status" json:"status"` // PENDING, CONFIRMED, TIMEOUT, CANCELLED
	LockID      string             `bson:"lock_id,omitempty" json:"lock_id,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	ConfirmedAt *time.Time         `bson:"confirmed_at,omitempty" json:"confirmed_at,omitempty"`
}

type User struct {
	ID           string   `bson:"_id" json:"id"`
	Email        string   `bson:"email" json:"email"`
	Name         string   `bson:"name" json:"name"`
	Role         UserRole `bson:"role" json:"role"`
	PasswordHash string   `bson:"password_hash,omitempty" json:"-"` // for email+password login
}

type AuditLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Event     string             `bson:"event" json:"event"`
	Payload   map[string]any     `bson:"payload" json:"payload"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

const (
	EventBookingSuccess  = "BOOKING_SUCCESS"
	EventBookingTimeout  = "BOOKING_TIMEOUT"
	EventSeatReleased    = "SEAT_RELEASED"
	EventSystemError     = "SYSTEM_ERROR"
	EventLockFailed      = "LOCK_FAIL"
)
