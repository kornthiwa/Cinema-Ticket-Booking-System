package mq

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
)

const (
	ChannelBookingEvents = "booking_events"
)

type Event struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}

type Publisher struct {
	client *redis.Client
}

func NewPublisher(client *redis.Client) *Publisher {
	return &Publisher{client: client}
}

func (p *Publisher) PublishBookingSuccess(ctx context.Context, screeningID, userID, bookingID string, seatRow, seatCol int) error {
	ev := Event{
		Type: "BOOKING_SUCCESS",
		Payload: map[string]any{
			"screening_id": screeningID,
			"user_id":      userID,
			"booking_id":   bookingID,
			"seat_row":     seatRow,
			"seat_col":     seatCol,
		},
	}
	return p.publish(ctx, ev)
}

func (p *Publisher) PublishSeatReleased(ctx context.Context, screeningID string, seatRow, seatCol int) error {
	ev := Event{
		Type: "SEAT_RELEASED",
		Payload: map[string]any{
			"screening_id": screeningID,
			"seat_row":     seatRow,
			"seat_col":     seatCol,
		},
	}
	return p.publish(ctx, ev)
}

func (p *Publisher) publish(ctx context.Context, ev Event) error {
	b, _ := json.Marshal(ev)
	return p.client.Publish(ctx, ChannelBookingEvents, b).Err()
}

type Subscriber struct {
	client    *redis.Client
	onEvent   func(Event)
}

func NewSubscriber(client *redis.Client, onEvent func(Event)) *Subscriber {
	return &Subscriber{client: client, onEvent: onEvent}
}

func (s *Subscriber) Run(ctx context.Context) {
	pubsub := s.client.Subscribe(ctx, ChannelBookingEvents)
	defer pubsub.Close()
	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			var ev Event
			if err := json.Unmarshal([]byte(msg.Payload), &ev); err != nil {
				log.Printf("mq: unmarshal event: %v", err)
				continue
			}
			s.onEvent(ev)
		}
	}
}
