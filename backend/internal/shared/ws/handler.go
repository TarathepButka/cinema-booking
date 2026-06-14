package ws

import (
	"net/http"

	"github.com/gin-gonic/gin"
	gws "github.com/gorilla/websocket"
)

type OriginValidator func(origin string) bool

// Handler upgrades WebSocket connections and registers clients with the hub.
type Handler struct {
	hub      *Hub
	upgrader gws.Upgrader
}

func NewHandler(hub *Hub, originAllowed OriginValidator) *Handler {
	return &Handler{
		hub: hub,
		upgrader: gws.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				return origin != "" && originAllowed(origin)
			},
		},
	}
}

// ServeWS handles GET /ws/showtimes/:id.
func (h *Handler) ServeWS(c *gin.Context) {
	showtimeID := c.Param("id")
	if showtimeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "showtime id required"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		ShowtimeID: showtimeID,
		Send:       make(chan []byte, 256),
		Hub:        h.hub,
	}

	h.hub.register <- client
	go client.writePump(conn)
	go client.readPump(conn)
}
