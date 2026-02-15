package handler

import (
	"log"
	"net/http"

	"cinema-booking/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ServeWS joins the client to room "screening:id" for real-time seat updates.
func (h *Handler) ServeWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}
	screeningID := c.Param("id")
	room := "screening:" + screeningID
	userID := c.GetString("user_id")
	client := &ws.Client{Send: make(chan []byte, 256), Room: room, UserID: userID}
	h.Hub.Register(client)
	defer h.Hub.Unregister(client)
	go func() {
		for b := range client.Send {
			if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
				break
			}
		}
	}()
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// ServeAdminWS joins the client to room "admin" for real-time refresh (bookings, audit logs).
func (h *Handler) ServeAdminWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("ws admin upgrade: %v", err)
		return
	}
	userID := c.GetString("user_id")
	client := &ws.Client{Send: make(chan []byte, 256), Room: ws.AdminRoom, UserID: userID}
	h.Hub.Register(client)
	defer h.Hub.Unregister(client)
	go func() {
		for b := range client.Send {
			if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
				break
			}
		}
	}()
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
