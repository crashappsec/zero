// Package ws provides WebSocket functionality for real-time updates
package ws

import (
	"context"
	"encoding/json"
	"sync"
)

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
