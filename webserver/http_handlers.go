package webserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"m32osc_controller/core" // Import core package

	"github.com/gorilla/mux"
)

// UpdateFaderRequest remains specific to webserver for decoding HTTP requests.
type UpdateFaderRequest struct {
	Level *float64 `json:"level,omitempty"`
	Muted *bool    `json:"muted,omitempty"`
}

// getFadersHandler now accepts *core.AppServer.
func getFadersHandler(app *core.AppServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		type FaderInfo struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
		}
		faderList := make([]FaderInfo, 0, len(app.Store))
		for _, fader := range app.Store { // app.Store contains *core.Fader
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
}

// getFaderHandler now accepts *core.AppServer.
func getFaderHandler(app *core.AppServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		faderID := strings.ToUpper(vars["id"])
		fader, ok := app.Store[faderID] // app.Store contains *core.Fader
		if !ok {
			log.Printf("Fader not found: %s", faderID)
			http.Error(w, fmt.Sprintf("Fader with ID '%s' not found", faderID), http.StatusNotFound)
			return
		}
		if err := json.NewEncoder(w).Encode(fader); err != nil { // fader is *core.Fader
			log.Printf("Error encoding fader %s: %v", faderID, err)
			http.Error(w, "Failed to encode fader data", http.StatusInternalServerError)
		}
	}
}

// updateFaderHandler now accepts *core.AppServer.
func updateFaderHandler(app *core.AppServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		faderID := strings.ToUpper(vars["id"])
		fader, ok := app.Store[faderID] // app.Store contains *core.Fader
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
			fader.Level = *req.Level // fader is *core.Fader
			log.Printf("Fader %s level updated to %f", faderID, fader.Level)
			updated = true
		}
		if req.Muted != nil {
			fader.Muted = *req.Muted // fader is *core.Fader
			log.Printf("Fader %s muted status updated to %t", faderID, fader.Muted)
			updated = true
		}
		if !updated {
			log.Printf("No update performed for fader %s as request body fields were empty or not provided.", faderID)
		} else {
			app.Events.Publish(fader) // app.Events is *core.Publisher, Publish takes *core.Fader
		}

		if err := json.NewEncoder(w).Encode(fader); err != nil { // fader is *core.Fader
			log.Printf("Error encoding updated fader %s: %v", faderID, err)
			http.Error(w, "Failed to encode fader data after update", http.StatusInternalServerError)
		}
	}
}
