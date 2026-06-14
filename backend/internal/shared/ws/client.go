package ws

import (
	"log"
	"time"

	gws "github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// readPump pumps messages from the WebSocket connection to the hub.
// Runs in a goroutine per client. Handles ping/pong heartbeat.
func (c *Client) readPump(conn *gws.Conn) {
	defer func() {
		c.Hub.unregister <- c
		conn.Close()
	}()

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// We don't expect meaningful messages from client → just drain
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			if gws.IsUnexpectedCloseError(err, gws.CloseGoingAway, gws.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection.
// Runs in a goroutine per client. Sends periodic pings.
func (c *Client) writePump(conn *gws.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				conn.WriteMessage(gws.CloseMessage, []byte{})
				return
			}
			w, err := conn.NextWriter(gws.TextMessage)
			if err != nil {
				return
			}
			w.Write(msg)
			// Flush remaining queued messages in same write frame
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.Send)
			}
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(gws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
