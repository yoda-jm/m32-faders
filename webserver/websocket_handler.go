package webserver

import (
	"log"
	"net/http"

	"m32osc_controller/core" // Import core package

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// serveWs now accepts *core.AppServer.
func serveWs(app *core.AppServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { // This is the main handler goroutine
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket: Failed to upgrade connection for client %s: %v", r.RemoteAddr, err)
			return
		}
		clientAddr := conn.RemoteAddr().String()
		log.Printf("WebSocket: Client %s connected.", clientAddr)

		subChan := make(core.Subscriber, 10) // Use core.Subscriber
		app.Events.Subscribe(subChan)
		log.Printf("WebSocket: Client %s subscribed to FaderEvents.", clientAddr)

		// This single defer in the main handler goroutine handles all cleanup.
		defer func() {
			log.Printf("WebSocket: Handler for client %s ending. Unsubscribing and closing connection.", clientAddr)
			app.Events.Unsubscribe(subChan)
			conn.Close()
		}()

		// Write Goroutine
		go func() {
			defer func() {
				if rcv := recover(); rcv != nil { // Changed variable name from r to rcv to avoid conflict
					log.Printf("WebSocket: Recovered from panic in write loop for client %s: %v", clientAddr, rcv)
				}
				log.Printf("WebSocket: Write loop for client %s ended.", clientAddr)
				// Note: No explicit Unsubscribe or conn.Close() here; main defer handles it.
				// If WriteJSON fails, it calls conn.Close() which terminates the read loop,
				// triggering the main defer.
			}()

			for update := range subChan { // update is of type core.FaderUpdate (*core.Fader)
				if update == nil {
					log.Printf("WebSocket: Received nil update for client %s. Skipping.", clientAddr)
					continue
				}
				log.Printf("WebSocket: Attempting to send update to client %s for fader %s: %+v", clientAddr, update.ID, update)
				if err := conn.WriteJSON(update); err != nil {
					log.Printf("WebSocket: Error writing JSON update to client %s: %v. Closing connection.", clientAddr, err)
					// Closing the connection here will cause the read loop in the main handler goroutine to fail and exit,
					// which in turn triggers the main defer cleanup.
					conn.Close() 
					return // Exit write goroutine
				}
			}
		}()

		// Read Loop (runs in the main handler goroutine, blocking it)
		log.Printf("WebSocket: Starting read loop for client %s.", clientAddr)
		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				// Log appropriately based on the type of error
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
					log.Printf("WebSocket: Error reading message from client %s (unexpected close): %v", clientAddr, err)
				} else {
					// This includes normal closures from client (e.g. browser tab closed) or "use of closed network connection"
					// if the write goroutine closed the connection due to a write error.
					log.Printf("WebSocket: Client %s disconnected or read error: %v", clientAddr, err)
				}
				break // Exit read loop, which will lead to deferred cleanup in this main handler goroutine.
			}
			log.Printf("WebSocket: Received message type %d from client %s: %s", messageType, clientAddr, string(p))
			// Process client messages here if needed in the future.
		}
		// When the read loop breaks, this function (the HTTP handler) will exit,
		// and its deferred cleanup will run.
	}
}
