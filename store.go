package main

import (
	"encoding/json" // Added for JSON parsing
	"fmt"
	"log"  // Added for logging
	"os"   // Added for file reading
)

// ConfigFaderType defines the structure for fader types in config.json
type ConfigFaderType struct {
	Type       string `json:"type"`
	Count      int    `json:"count"`
	IDPrefix   string `json:"idPrefix"`
	NamePrefix string `json:"namePrefix"`
}

// ConfigMasterFader defines the structure for master faders in config.json
type ConfigMasterFader struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AppConfig defines the overall structure of config.json
type AppConfig struct {
	FaderTypes     []ConfigFaderType   `json:"faderTypes"`
	MasterFaders   []ConfigMasterFader `json:"masterFaders"`
	DefaultLevel float64             `json:"defaultLevel"`
	DefaultMuted bool                `json:"defaultMuted"`
}

// MasterFaderStore holds all the faders available on the console.
// The key is the Fader.ID.
var MasterFaderStore map[string]*Fader

func init() {
	MasterFaderStore = make(map[string]*Fader)

	// Read config.json
	configFile := "config.json"
	fileBytes, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error reading config file '%s': %v", configFile, err)
	}

	// Parse JSON
	var appConfig AppConfig
	err = json.Unmarshal(fileBytes, &appConfig)
	if err != nil {
		log.Fatalf("Error parsing config file '%s': %v", configFile, err)
	}

	log.Printf("Successfully loaded configuration from %s", configFile)

	// Populate MasterFaderStore from appConfig.FaderTypes
	for _, ft := range appConfig.FaderTypes {
		log.Printf("Loading fader type: %s (Count: %d)", ft.Type, ft.Count)
		for i := 1; i <= ft.Count; i++ {
			id := fmt.Sprintf("%s%02d", ft.IDPrefix, i)
			// Ensure ID is uppercase for consistency with how it might be looked up (e.g. from URL params)
			// id = strings.ToUpper(id) // Decided against this for now, to match exact config.json format.
			name := fmt.Sprintf("%s%d", ft.NamePrefix, i)

			MasterFaderStore[id] = &Fader{
				ID:    id,
				Name:  name,
				Type:  ft.Type,
				Level: appConfig.DefaultLevel,
				Muted: appConfig.DefaultMuted,
			}
		}
	}

	// Populate MasterFaderStore from appConfig.MasterFaders
	for _, mf := range appConfig.MasterFaders {
		log.Printf("Loading master fader: %s (ID: %s)", mf.Name, mf.ID)
		// id := strings.ToUpper(mf.ID) // Consistent ID handling
		id := mf.ID
		MasterFaderStore[id] = &Fader{
			ID:    id,
			Name:  mf.Name,
			Type:  mf.Type,
			Level: appConfig.DefaultLevel,
			Muted: appConfig.DefaultMuted,
		}
	}
	log.Printf("MasterFaderStore initialized with %d faders from config.", len(MasterFaderStore))
}
