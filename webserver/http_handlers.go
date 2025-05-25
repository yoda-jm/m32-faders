package webserver // Changed package name

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"m32osc_controller/main" // To access main.AppServer (assuming AppServer is in package main)

	"github.com/gorilla/mux"
)

// UpdateFaderRequest is used to decode partial updates for a fader.
// Using pointers to distinguish between a field not being present and being set to its zero value.
// This struct remains in the webserver package as it's specific to the HTTP update handler.
type UpdateFaderRequest struct {
	Level *float64 `json:"level,omitempty"`
	Muted *bool    `json:"muted,omitempty"`
}

// getFadersHandler returns an http.HandlerFunc that lists faders.
// It now accepts an AppServer instance from the main package.
func getFadersHandler(app *main.AppServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		type FaderInfo struct { // This local struct is fine
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
		}
		faderList := make([]FaderInfo, 0, len(app.Store))
		for _, fader := range app.Store { // app.Store refers to main.AppServer.Store
			faderList = append(faderList, FaderInfo{
				ID:   fader.ID,   // fader is *main.Fader, fields are public
				Name: fader.Name,
				Type: fader.Type,
			})
		}
		if err := json.NewEncoder(w).Encode(faderList); err != nil {
			log.Printf("Error encoding faders list: %v", err)
			http.Error(w, "Failed to encode fader list", http.StatusInternalServerError)
		}
	}
}

// getFaderHandler returns an http.HandlerFunc that gets a specific fader.
// It now accepts an AppServer instance from the main package.
func getFaderHandler(app *main.AppServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		faderID := strings.ToUpper(vars["id"])
		fader, ok := app.Store[faderID] // app.Store refers to main.AppServer.Store
		if !ok {
			log.Printf("Fader not found: %s", faderID)
			http.Error(w, fmt.Sprintf("Fader with ID '%s' not found", faderID), http.StatusNotFound)
			return
		}
		if err := json.NewEncoder(w).Encode(fader); err != nil { // fader is *main.Fader
			log.Printf("Error encoding fader %s: %v", faderID, err)
			http.Error(w, "Failed to encode fader data", http.StatusInternalServerError)
		}
	}
}

// updateFaderHandler returns an http.HandlerFunc that updates a specific fader.
// It now accepts an AppServer instance from the main package.
func updateFaderHandler(app *main.AppServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		faderID := strings.ToUpper(vars["id"])
		fader, ok := app.Store[faderID] // app.Store refers to main.AppServer.Store
		if !ok {
			log.Printf("Fader not found for update: %s", faderID)
			http.Error(w, fmt.Sprintf("Fader with ID '%s' not found", faderID), http.StatusNotFound)
			return
		}
		var req UpdateFaderRequest // This uses the local webserver.UpdateFaderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Error decoding update request for fader %s: %v", faderID, err)
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}
		updated := false
		if req.Level != nil {
			fader.Level = *req.Level // fader is *main.Fader
			log.Printf("Fader %s level updated to %f", faderID, fader.Level)
			updated = true
		}
		if req.Muted != nil {
			fader.Muted = *req.Muted // fader is *main.Fader
			log.Printf("Fader %s muted status updated to %t", faderID, fader.Muted)
			updated = true
		}
		if !updated {
			log.Printf("No update performed for fader %s as request body fields were empty or not provided.", faderID)
		} else {
			app.Events.Publish(fader) // app.Events refers to main.AppServer.Events, Publish takes *main.Fader
		}

		if err := json.NewEncoder(w).Encode(fader); err != nil { // fader is *main.Fader
			log.Printf("Error encoding updated fader %s: %v", faderID, err)
			http.Error(w, "Failed to encode fader data after update", http.StatusInternalServerError)
		}
	}
}
