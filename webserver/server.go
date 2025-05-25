package webserver

import (
	"log"
	"net/http"
	"path/filepath" // For static files

	"m32osc_controller/main" // To access main.AppServer

	"github.com/gorilla/mux"
)

// loggingMiddleware logs incoming HTTP requests.
// It's an unexported function, local to the webserver package.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received HTTP request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// Start initializes and starts the HTTP server.
// It takes the listen address and the main application server instance.
func Start(httpAddr string, app *main.AppServer) {
	r := mux.NewRouter()

	// Middleware for logging requests
	r.Use(loggingMiddleware)

	// API routes
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/faders", getFadersHandler(app)).Methods("GET")       // getFadersHandler is from http_handlers.go in this package
	apiRouter.HandleFunc("/faders/{id}", getFaderHandler(app)).Methods("GET")   // getFaderHandler is from http_handlers.go
	apiRouter.HandleFunc("/faders/{id}", updateFaderHandler(app)).Methods("POST") // updateFaderHandler is from http_handlers.go

	// Static file serving for assets (e.g., JS, CSS)
	// Ensure path is relative to where binary runs (project root)
	jsFileServer := http.FileServer(http.Dir("./static/js/"))
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", jsFileServer))

	cssFileServer := http.FileServer(http.Dir("./static/css/"))
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", cssFileServer))

	// Route for index.html at the root
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "index.html"))
	}).Methods("GET")

	// WebSocket endpoint
	r.HandleFunc("/ws", serveWs(app)) // serveWs is from websocket_handler.go in this package

	log.Printf("HTTP server (webserver package) starting on %s", httpAddr)
	if err := http.ListenAndServe(httpAddr, r); err != nil {
		// Use log.Fatalf as this is a critical error for the HTTP server.
		// The main application might still run the TCP server, but the HTTP part is down.
		// Or, if HTTP is essential, this could be propagated up. For now, Fatalf.
		log.Fatalf("webserver.Start: ListenAndServe failed: %v", err)
	}
}
