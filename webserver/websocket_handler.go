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
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket: Failed to upgrade connection for client %s: %v", r.RemoteAddr, err)
			return
		}
		clientAddr := conn.RemoteAddr().String() // Get client address string once
		log.Printf("WebSocket: Client %s connected.", clientAddr)

		defer func() {
			log.Printf("WebSocket: Closing connection for client %s.", clientAddr)
			conn.Close()
		}()

		subChan := make(core.Subscriber, 10) // Use core.Subscriber
		app.Events.Subscribe(subChan)        // app.Events is *core.Publisher
		log.Printf("WebSocket: Client %s subscribed to FaderEvents.", clientAddr)

		// Goroutine to send fader updates to the client
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("WebSocket: Recovered from panic in write loop for client %s: %v", clientAddr, r)
				}
				log.Printf("WebSocket: Write loop for client %s ended. Unsubscribing.", clientAddr)
				app.Events.Unsubscribe(subChan) // app.Events is *core.Publisher
			}()

			for update := range subChan { // update is of type core.FaderUpdate (*core.Fader)
				if update == nil {
					log.Printf("WebSocket: Received nil update for client %s. Skipping.", clientAddr)
					continue
				}
				// Added detailed log before sending
				log.Printf("WebSocket: Attempting to send update to client %s for fader %s: %+v", clientAddr, update.ID, update)
				err := conn.WriteJSON(update) // update is *core.Fader
				if err != nil {
					log.Printf("WebSocket: Error writing JSON update to client %s: %v", clientAddr, err)
					return // Exit goroutine, defer will handle unsubscription
				}
			}
			log.Printf("WebSocket: FaderEvents channel closed for client %s. Write loop exiting.", clientAddr)
		}()

		// Goroutine to read messages from the client (primarily to detect disconnections)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("WebSocket: Recovered from panic in read loop for client %s: %v", clientAddr, r)
				}
				log.Printf("WebSocket: Read loop for client %s ended. Unsubscribing and closing connection.", clientAddr)
				app.Events.Unsubscribe(subChan) // app.Events is *core.Publisher
				conn.Close()                     // Ensure connection is closed if read loop exits
			}()

			for {
				messageType, p, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
						log.Printf("WebSocket: Error reading message from client %s (unexpected close): %v", clientAddr, err)
					} else {
						// Normal closure (e.g., client called socket.close() or browser tab closed)
						// or other read error that implies disconnection.
						log.Printf("WebSocket: Client %s disconnected or read error: %v", clientAddr, err)
					}
					break // Exit loop on any error, triggering deferred cleanup
				}
				log.Printf("WebSocket: Received message type %d from client %s: %s", messageType, clientAddr, string(p))
				// No specific client messages are handled in this application yet.
			}
		}()
	}
}
