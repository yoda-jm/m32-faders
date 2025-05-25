package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections for now.
		// For production, you might want to check r.Header.Get("Origin")
		// against a list of allowed origins.
		return true
	},
}

// serveWs handles WebSocket requests from clients.
func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket: Failed to upgrade connection for client %s: %v", r.RemoteAddr, err)
		return
	}
	clientAddr := conn.RemoteAddr().String()
	log.Printf("WebSocket: Client %s connected.", clientAddr)

	// Ensure connection is closed and unsubscribed when function exits
	// The order of defer matters: Unsubscribe first, then close connection.
	defer func() {
		log.Printf("WebSocket: Closing connection and unsubscribing client %s.", clientAddr)
		// Unsubscribe logic will be handled by the read/write loops triggering closure.
		// Explicitly calling conn.Close() here ensures it's closed if serveWs exits for other reasons.
		conn.Close()
	}()

	subChan := make(Subscriber, 10) // Buffered channel for fader updates
	FaderEvents.Subscribe(subChan)
	log.Printf("WebSocket: Client %s subscribed to FaderEvents.", clientAddr)

	// Goroutine to send fader updates to the client
	go func() {
		// Ensure unsubscription when this goroutine exits (e.g., due to write error)
		defer func() {
			log.Printf("WebSocket: Write loop for client %s ended. Unsubscribing.", clientAddr)
			FaderEvents.Unsubscribe(subChan)
			// conn.Close() // Closing the connection here might be redundant if read loop also closes it
		}()

		for update := range subChan { // This loop breaks when subChan is closed
			if update == nil { // Should not happen if FaderUpdate is *Fader, but good check
				log.Printf("WebSocket: Received nil update for client %s. Skipping.", clientAddr)
				continue
			}
			// log.Printf("WebSocket: Sending update for fader %s to client %s", update.ID, clientAddr)
			err := conn.WriteJSON(update) // update is already a *Fader
			if err != nil {
				log.Printf("WebSocket: Error writing JSON update to client %s: %v", clientAddr, err)
				// Assume client disconnected, break loop to trigger unsubscribe and connection close.
				return
			}
		}
		log.Printf("WebSocket: FaderEvents channel closed for client %s. Write loop exiting.", clientAddr)
	}()

	// Goroutine to read messages from the client (primarily to detect disconnections)
	go func() {
		// Ensure unsubscription and connection close when this goroutine exits
		defer func() {
			log.Printf("WebSocket: Read loop for client %s ended. Unsubscribing and closing connection.", clientAddr)
			FaderEvents.Unsubscribe(subChan) // Critical to unsubscribe
			conn.Close()                     // Ensure connection is closed
		}()

		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
					log.Printf("WebSocket: Error reading message from client %s (unexpected close): %v", clientAddr, err)
				} else {
					// Normal closure or known error (e.g. client closed connection, browser tab closed)
					log.Printf("WebSocket: Client %s disconnected or read error: %v", clientAddr, err)
				}
				break // Exit loop on any error, triggering deferred cleanup
			}
			// Log received messages (optional, good for debugging)
			log.Printf("WebSocket: Received message type %d from client %s: %s", messageType, clientAddr, string(p))

			// Here you could handle control messages like Ping/Pong if needed,
			// or specific application messages from the client if your protocol defines them.
		}
	}()
}
