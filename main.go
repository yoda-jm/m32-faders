package main

import (
	"flag" // Added for command-line flags
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

// UpdateFaderRequest and handlers (getFadersHandler, getFaderHandler, updateFaderHandler)
// have been moved to http_handlers.go

func main() {
	// Define command-line flags for server addresses
	httpAddr := flag.String("http-addr", "localhost:8080", "HTTP server address (e.g., localhost:8080 or :8080)")
	tcpAddr := flag.String("tcp-addr", "localhost:8001", "TCP server address (e.g., localhost:8001 or :8001)")
	flag.Parse() // Parse the command-line flags

	r := mux.NewRouter()

	// Middleware for logging requests
	r.Use(loggingMiddleware)

	// API routes (defined first for precedence)
	// These handlers are now defined in http_handlers.go
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/faders", getFadersHandler).Methods("GET")
	apiRouter.HandleFunc("/faders/{id}", getFaderHandler).Methods("GET")
	apiRouter.HandleFunc("/faders/{id}", updateFaderHandler).Methods("POST")

	// Static file serving for assets (e.g., JS, CSS)
	jsFileServer := http.FileServer(http.Dir("./static/js/"))
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", jsFileServer))

	cssFileServer := http.FileServer(http.Dir("./static/css/"))
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", cssFileServer))

	// Route for index.html at the root
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "index.html"))
	}).Methods("GET")

	// WebSocket endpoint
	r.HandleFunc("/ws", serveWs)

	// Start the TCP server in a goroutine using the address from the flag
	go StartTCPServer(*tcpAddr)

	log.Printf("M32 Fader Control HTTP server starting on %s (serving API, static files, and WebSocket)", *httpAddr)
	// Use the address from the flag for the HTTP server
	log.Fatal(http.ListenAndServe(*httpAddr, r))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received HTTP request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
