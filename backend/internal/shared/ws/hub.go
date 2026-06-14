package ws

import (
	"log"
	"sync"
)

const MessageTypeSeatUpdate = "SEAT_UPDATE"

// SeatUpdateMessage is the message broadcast to all clients in a showtime room.
type SeatUpdateMessage struct {
	Type       string `json:"type"` // "SEAT_UPDATE"
	ShowtimeID string `json:"showtimeId"`
	SeatLabel  string `json:"seatLabel"`
	Status     string `json:"status"` // AVAILABLE, LOCKED, BOOKED
	UpdatedBy  string `json:"updatedBy,omitempty"`
}

// Client represents a single WebSocket connection.
type Client struct {
	ShowtimeID string
	Send       chan []byte
	Hub        *Hub
}

// Hub manages all active WebSocket connections, organized by showtime room.
type Hub struct {
	// rooms: showtimeID → set of clients
	rooms map[string]map[*Client]bool
	mu    sync.RWMutex

	// Channels for registration and broadcast
	register   chan *Client
	unregister chan *Client
	broadcast  chan broadcastMsg
}

type broadcastMsg struct {
	showtimeID string
	data       []byte
}

// NewHub creates and returns a new Hub. Call Run() to start it.
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		broadcast:  make(chan broadcastMsg, 256),
	}
}

// Run starts the hub's event loop. Must be called in a goroutine.
func (h *Hub) Run() {
	log.Println("🔌 WebSocket Hub running")
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.rooms[client.ShowtimeID] == nil {
				h.rooms[client.ShowtimeID] = make(map[*Client]bool)
			}
			h.rooms[client.ShowtimeID][client] = true
			count := len(h.rooms[client.ShowtimeID])
			h.mu.Unlock()
			log.Printf("🔌 Client joined room %s (total: %d)", client.ShowtimeID, count)

		case client := <-h.unregister:
			h.mu.Lock()
			if room, ok := h.rooms[client.ShowtimeID]; ok {
				if _, ok := room[client]; ok {
					delete(room, client)
					close(client.Send)
					if len(room) == 0 {
						delete(h.rooms, client.ShowtimeID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("🔌 Client left room %s", client.ShowtimeID)

		case msg := <-h.broadcast:
			h.mu.RLock()
			clients := h.rooms[msg.showtimeID]
			h.mu.RUnlock()

			for client := range clients {
				select {
				case client.Send <- msg.data:
				default:
					// Slow client — disconnect
					h.mu.Lock()
					delete(h.rooms[client.ShowtimeID], client)
					close(client.Send)
					h.mu.Unlock()
				}
			}
		}
	}
}

// Broadcast sends a message to all clients watching a showtime.
func (h *Hub) Broadcast(showtimeID string, data []byte) {
	h.broadcast <- broadcastMsg{showtimeID: showtimeID, data: data}
}

// ClientCount returns the number of connected clients for a showtime.
func (h *Hub) ClientCount(showtimeID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms[showtimeID])
}
