package webserver

import (
	"log"
	"net/http"
	"path/filepath" // For static files

	"m32osc_controller/core" // Import core package

	"github.com/gorilla/mux"
)

// loggingMiddleware logs incoming HTTP requests.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received HTTP request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// Start initializes and starts the HTTP server.
// It now accepts a *core.AppServer instance.
func Start(httpAddr string, app *core.AppServer) {
	r := mux.NewRouter()

	r.Use(loggingMiddleware)

	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/faders", getFadersHandler(app)).Methods("GET")
	apiRouter.HandleFunc("/faders/{id}", getFaderHandler(app)).Methods("GET")
	apiRouter.HandleFunc("/faders/{id}", updateFaderHandler(app)).Methods("POST")

	jsFileServer := http.FileServer(http.Dir("./static/js/"))
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", jsFileServer))

	cssFileServer := http.FileServer(http.Dir("./static/css/"))
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", cssFileServer))

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "index.html"))
	}).Methods("GET")

	r.HandleFunc("/ws", serveWs(app))

	log.Printf("HTTP server (webserver package) starting on %s", httpAddr)
	if err := http.ListenAndServe(httpAddr, r); err != nil {
		log.Fatalf("webserver.Start: ListenAndServe failed: %v", err)
	}
}
