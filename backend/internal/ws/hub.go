package ws

import (
	"encoding/json"
	"log"
	"sync"
)

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Client struct {
	Send   chan []byte
	Room   string
	UserID string
}

type Hub struct {
	mu      sync.RWMutex
	rooms   map[string]map[*Client]struct{}
	broadcast chan roomMessage
	register  chan *Client
	unregister chan *Client
}

type roomMessage struct {
	Room string
	Msg  []byte
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]struct{}),
		broadcast:  make(chan roomMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			if h.rooms[c.Room] == nil {
				h.rooms[c.Room] = make(map[*Client]struct{})
			}
			h.rooms[c.Room][c] = struct{}{}
			h.mu.Unlock()

		case c := <-h.unregister:
			h.mu.Lock()
			if m, ok := h.rooms[c.Room]; ok {
				delete(m, c)
				if len(m) == 0 {
					delete(h.rooms, c.Room)
				}
			}
			close(c.Send)
			h.mu.Unlock()

		case rm := <-h.broadcast:
			h.mu.RLock()
			for c := range h.rooms[rm.Room] {
				select {
				case c.Send <- rm.Msg:
				default:
					// skip slow client
				}
			}
			h.mu.RUnlock()
		}
	}
}

const AdminRoom = "admin"

func (h *Hub) BroadcastSeatUpdate(room string, payload interface{}) {
	msg := Message{Type: "SEAT_UPDATE", Payload: payload}
	b, err := json.Marshal(msg)
	if err != nil {
		log.Printf("ws: marshal: %v", err)
		return
	}
	select {
	case h.broadcast <- roomMessage{Room: room, Msg: b}:
	default:
		log.Printf("ws: broadcast buffer full for room %s", room)
	}
}

// BroadcastAdmin sends a message to all clients in the admin room (e.g. REFRESH for live bookings/audit).
func (h *Hub) BroadcastAdmin(msgType string, payload interface{}) {
	msg := Message{Type: msgType, Payload: payload}
	b, err := json.Marshal(msg)
	if err != nil {
		log.Printf("ws: marshal admin: %v", err)
		return
	}
	select {
	case h.broadcast <- roomMessage{Room: AdminRoom, Msg: b}:
	default:
		log.Printf("ws: broadcast buffer full for room %s", AdminRoom)
	}
}

// BroadcastNotification sends a NOTIFICATION message to a room (e.g. screening:<id>) for frontend to show toast.
func (h *Hub) BroadcastNotification(room string, eventType string, payload interface{}) {
	msg := Message{
		Type: "NOTIFICATION",
		Payload: map[string]interface{}{
			"eventType": eventType,
			"payload":   payload,
		},
	}
	b, err := json.Marshal(msg)
	if err != nil {
		log.Printf("ws: marshal notification: %v", err)
		return
	}
	select {
	case h.broadcast <- roomMessage{Room: room, Msg: b}:
	default:
		log.Printf("ws: broadcast buffer full for room %s", room)
	}
}

func (h *Hub) Register(c *Client) { h.register <- c }
func (h *Hub) Unregister(c *Client) { h.unregister <- c }
