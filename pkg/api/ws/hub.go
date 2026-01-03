// Package ws provides WebSocket functionality for real-time updates
package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, validate origin properly
		return true
	},
}

// Hub maintains active WebSocket connections and broadcasts messages
type Hub struct {
	// Registered clients by topic
	clients map[string]map[*Client]bool

	// Register requests
	register chan *registration

	// Unregister requests
	unregister chan *Client

	// Broadcast messages
	broadcast chan *broadcastMessage

	mu sync.RWMutex
}

// Client represents a WebSocket client connection
type Client struct {
	hub   *Hub
	conn  *websocket.Conn
	topic string
	send  chan []byte
}

type registration struct {
	client *Client
	topic  string
}

type broadcastMessage struct {
	topic string
	data  []byte
}

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *registration),
		unregister: make(chan *Client),
		broadcast:  make(chan *broadcastMessage, 256),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case reg := <-h.register:
			h.mu.Lock()
			if h.clients[reg.topic] == nil {
				h.clients[reg.topic] = make(map[*Client]bool)
			}
			h.clients[reg.topic][reg.client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.topic]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.clients, client.topic)
					}
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.clients[msg.topic]; ok {
				for client := range clients {
					select {
					case client.send <- msg.data:
					default:
						// Client buffer full, schedule removal
						go func(c *Client) {
							h.unregister <- c
						}(client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all clients subscribed to a topic
func (h *Hub) Broadcast(topic string, msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	h.broadcast <- &broadcastMessage{
		topic: topic,
		data:  data,
	}
	return nil
}

// BroadcastToJob sends a message to all clients watching a specific job
func (h *Hub) BroadcastToJob(jobID string, msg interface{}) error {
	return h.Broadcast("job:"+jobID, msg)
}

// BroadcastToAgent sends a message to a specific agent chat session
func (h *Hub) BroadcastToAgent(sessionID string, msg interface{}) error {
	return h.Broadcast("agent:"+sessionID, msg)
}

// Register adds a client to a topic
func (h *Hub) Register(client *Client, topic string) {
	h.register <- &registration{
		client: client,
		topic:  topic,
	}
}

// Unregister removes a client
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// ClientCount returns the number of clients for a topic
func (h *Hub) ClientCount(topic string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[topic])
}

// HandleScanWS handles WebSocket connections for scan job updates
func (h *Hub) HandleScanWS(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobID")
	if jobID == "" {
		http.Error(w, "job ID required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		hub:   h,
		conn:  conn,
		topic: "job:" + jobID,
		send:  make(chan []byte, 256),
	}

	h.register <- &registration{client: client, topic: client.topic}

	// Send initial connection confirmation
	conn.WriteJSON(Message{
		Type: "connected",
		Payload: map[string]string{
			"job_id": jobID,
			"topic":  client.topic,
		},
	})

	// Start read and write pumps
	go client.writePump()
	go client.readPump()
}

// HandleAgentWS handles WebSocket connections for agent chat
func (h *Hub) HandleAgentWS(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		sessionID = "default"
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		hub:   h,
		conn:  conn,
		topic: "agent:" + sessionID,
		send:  make(chan []byte, 256),
	}

	h.register <- &registration{client: client, topic: client.topic}

	// Send initial connection confirmation
	conn.WriteJSON(Message{
		Type: "connected",
		Payload: map[string]string{
			"session_id": sessionID,
			"topic":      client.topic,
		},
	})

	// Start read and write pumps
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Process incoming messages (for future bidirectional communication)
		var msg Message
		if err := json.Unmarshal(message, &msg); err == nil {
			// Handle client messages if needed
			switch msg.Type {
			case "ping":
				c.send <- []byte(`{"type":"pong"}`)
			}
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Drain queued messages to the current WebSocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
