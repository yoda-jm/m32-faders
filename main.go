package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// UpdateFaderRequest is used to decode partial updates for a fader.
// Using pointers to distinguish between a field not being present and being set to its zero value.
type UpdateFaderRequest struct {
	Level *float64 `json:"level,omitempty"`
	Muted *bool    `json:"muted,omitempty"`
}

func main() {
	r := mux.NewRouter()

	// Middleware for logging requests
	r.Use(loggingMiddleware)

	// Routes
	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/api/faders", getFadersHandler).Methods("GET")
	r.HandleFunc("/api/faders/{id}", getFaderHandler).Methods("GET")
	r.HandleFunc("/api/faders/{id}", updateFaderHandler).Methods("POST")

	// Start the TCP server in a goroutine
	go StartTCPServer("8001")

	log.Println("M32 Fader Control API server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "M32 Fader Control API Ready"})
}

func getFadersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Create a slice of simplified fader info for the response
	type FaderInfo struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	}

	faderList := make([]FaderInfo, 0, len(MasterFaderStore))
	for _, fader := range MasterFaderStore {
		faderList = append(faderList, FaderInfo{
			ID:   fader.ID,
			Name: fader.Name,
			Type: fader.Type,
		})
	}

	if err := json.NewEncoder(w).Encode(faderList); err != nil {
		log.Printf("Error encoding faders list: %v", err)
		http.Error(w, "Failed to encode fader list", http.StatusInternalServerError)
	}
}

func getFaderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	faderID := strings.ToUpper(vars["id"]) // Ensure ID is uppercase to match store keys

	fader, ok := MasterFaderStore[faderID]
	if !ok {
		log.Printf("Fader not found: %s", faderID)
		http.Error(w, fmt.Sprintf("Fader with ID '%s' not found", faderID), http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(fader); err != nil {
		log.Printf("Error encoding fader %s: %v", faderID, err)
		http.Error(w, "Failed to encode fader data", http.StatusInternalServerError)
	}
}

func updateFaderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	faderID := strings.ToUpper(vars["id"]) // Ensure ID is uppercase

	fader, ok := MasterFaderStore[faderID]
	if !ok {
		log.Printf("Fader not found for update: %s", faderID)
		http.Error(w, fmt.Sprintf("Fader with ID '%s' not found", faderID), http.StatusNotFound)
		return
	}

	var req UpdateFaderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding update request for fader %s: %v", faderID, err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	updated := false
	if req.Level != nil {
		fader.Level = *req.Level
		log.Printf("Fader %s level updated to %f", faderID, fader.Level)
		updated = true
	}
	if req.Muted != nil {
		fader.Muted = *req.Muted
		log.Printf("Fader %s muted status updated to %t", faderID, fader.Muted)
		updated = true
	}

	if !updated {
		log.Printf("No update performed for fader %s as request body fields were empty or not provided.", faderID)
		// It's debatable whether to return an error or just the current state.
		// For now, returning current state. Could also be a 304 Not Modified if nothing changed.
	}
	
	// Persist changes (if this were a real database)
	// MasterFaderStore[faderID] = fader // In our case, fader is a pointer, so modifications are direct.

	if err := json.NewEncoder(w).Encode(fader); err != nil {
		log.Printf("Error encoding updated fader %s: %v", faderID, err)
		http.Error(w, "Failed to encode fader data after update", http.StatusInternalServerError)
	}
}
