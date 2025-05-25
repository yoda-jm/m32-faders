package main

import (
	"flag"
	"log"
	// "net/http" // No longer needed directly for ListenAndServe or middleware
	// "path/filepath" // No longer needed for static file serving here
	// "github.com/gorilla/mux" // No longer needed for router setup here

	"m32osc_controller/webserver" // Import the new webserver package
)

// Types (Fader, AppServer, Publisher, Subscriber) remain in package main.
// UpdateFaderRequest is in webserver package.
// HTTP handlers and WebSocket handler are now in webserver package.

func main() {
	// Define command-line flags for server addresses
	httpAddr := flag.String("http-addr", "localhost:8080", "HTTP server address (e.g., localhost:8080 or :8080)")
	tcpAddr := flag.String("tcp-addr", "localhost:8001", "TCP server address (e.g., localhost:8001 or :8001)")
	configPath := flag.String("config", "config.json", "Path to the configuration file")
	flag.Parse() // Parse the command-line flags

	// Create the AppServer instance
	app, err := NewAppServer(*configPath) // NewAppServer returns (*AppServer, error)
	if err != nil {
		log.Fatalf("Failed to initialize AppServer: %v", err)
	}

	// Start the TCP server in a goroutine using the address from the flag
	// StartTCPServer now accepts 'app'.
	go StartTCPServer(*tcpAddr, app)

	// Start the OSC client placeholder
	go StartOSCClient(app)

	// Start the HTTP server (this is a blocking call)
	// The webserver.Start function now handles router setup, middleware, and ListenAndServe.
	log.Printf("Starting HTTP server on %s via webserver package", *httpAddr)
	webserver.Start(*httpAddr, app) // This will block and log if ListenAndServe fails
}

// loggingMiddleware has been moved to webserver/server.go
// func loggingMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		log.Printf("Received HTTP request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
// 		next.ServeHTTP(w, r)
// 	})
// }
