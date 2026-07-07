// Herlin AI Assistant - Backend Service
// Copyright 2026 Herlin AI. All rights reserved.

package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

type Message struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	UserID    uint        `json:"user_id,omitempty"`
	ChatID    uint        `json:"chat_id,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

type Client struct {
	ID     string
	UserID uint
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan Message
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("Client connected: %s (User ID: %d)", client.ID, client.UserID)

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mutex.Unlock()
			log.Printf("Client disconnected: %s (User ID: %d)", client.ID, client.UserID)

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				// Send to all clients or specific user/chat
				if message.UserID == 0 || client.UserID == message.UserID {
					select {
					case client.Send <- message:
					default:
						close(client.Send)
						delete(h.clients, client)
					}
				}
			}
			h.mutex.RUnlock()
		}
	}
}

func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

func (h *Hub) Broadcast(message Message) {
	h.broadcast <- message
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.UnregisterClient(c)
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("JSON unmarshal error: %v", err)
			continue
		}

		msg.UserID = c.UserID
		msg.Timestamp = time.Now().Unix()

		// Handle different message types
		switch msg.Type {
		case "chat":
			// Broadcast chat message
			c.Hub.Broadcast(msg)
		case "typing":
			// Broadcast typing indicator
			c.Hub.Broadcast(msg)
		}
	}
}

func (c *Client) WritePump() {
	defer c.Conn.Close()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}

func HandleWebSocket(hub *Hub, c *gin.Context) {
	userID := c.GetUint("user_id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		ID:     generateClientID(),
		UserID: userID,
		Hub:    hub,
		Conn:   conn,
		Send:   make(chan Message, 256),
	}

	hub.RegisterClient(client)

	// Start pumps
	go client.WritePump()
	client.ReadPump()
}

func generateClientID() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}
