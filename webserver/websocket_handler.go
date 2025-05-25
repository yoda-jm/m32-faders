package webserver // Changed package name

import (
	"log"
	"net/http"

	"m32osc_controller/main" // To access main.AppServer and main.Subscriber

	"github.com/gorilla/websocket"
)

// upgrader remains a package-level variable within webserver
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// serveWs returns an http.HandlerFunc that handles WebSocket requests from clients.
// It now accepts an AppServer instance from the main package.
func serveWs(app *main.AppServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket: Failed to upgrade connection for client %s: %v", r.RemoteAddr, err)
			return
		}
		clientAddr := conn.RemoteAddr().String()
		log.Printf("WebSocket: Client %s connected.", clientAddr)

		defer func() {
			log.Printf("WebSocket: Closing connection for client %s.", clientAddr)
			conn.Close()
		}()

		subChan := make(main.Subscriber, 10) // Use main.Subscriber
		app.Events.Subscribe(subChan)        // app.Events refers to main.AppServer.Events
		log.Printf("WebSocket: Client %s subscribed to FaderEvents.", clientAddr)

		// Goroutine to send fader updates to the client
		go func() {
			defer func() {
				log.Printf("WebSocket: Write loop for client %s ended. Unsubscribing.", clientAddr)
				app.Events.Unsubscribe(subChan) // Use main.AppServer.Events
			}()

			for update := range subChan { // update is of type main.FaderUpdate (which is *main.Fader)
				if update == nil {
					log.Printf("WebSocket: Received nil update for client %s. Skipping.", clientAddr)
					continue
				}
				err := conn.WriteJSON(update) // update is *main.Fader
				if err != nil {
					log.Printf("WebSocket: Error writing JSON update to client %s: %v", clientAddr, err)
					return
				}
			}
			log.Printf("WebSocket: FaderEvents channel closed for client %s. Write loop exiting.", clientAddr)
		}()

		// Goroutine to read messages from the client (primarily to detect disconnections)
		go func() {
			defer func() {
				log.Printf("WebSocket: Read loop for client %s ended. Unsubscribing and closing connection.", clientAddr)
				app.Events.Unsubscribe(subChan) // Use main.AppServer.Events
				conn.Close()
			}()

			for {
				messageType, p, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
						log.Printf("WebSocket: Error reading message from client %s (unexpected close): %v", clientAddr, err)
					} else {
						log.Printf("WebSocket: Client %s disconnected or read error: %v", clientAddr, err)
					}
					break
				}
				log.Printf("WebSocket: Received message type %d from client %s: %s", messageType, clientAddr, string(p))
			}
		}()
	}
}
