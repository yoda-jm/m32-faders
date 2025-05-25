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

// MasterFaderStore is no longer a global variable.
// var MasterFaderStore map[string]*Fader

// NewFaderStore creates and returns a new fader store, initialized from the given config file.
func NewFaderStore(configPath string) (map[string]*Fader, error) {
	store := make(map[string]*Fader)

	// Read config.json
	fileBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file '%s': %w", configPath, err)
	}

	// Parse JSON
	var appConfig AppConfig
	err = json.Unmarshal(fileBytes, &appConfig)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file '%s': %w", configPath, err)
	}

	log.Printf("Successfully loaded configuration from %s", configPath)

	// Populate store from appConfig.FaderTypes
	for _, ft := range appConfig.FaderTypes {
		log.Printf("Loading fader type: %s (Count: %d)", ft.Type, ft.Count)
		for i := 1; i <= ft.Count; i++ {
			id := fmt.Sprintf("%s%02d", ft.IDPrefix, i)
			name := fmt.Sprintf("%s%d", ft.NamePrefix, i)

			store[id] = &Fader{
				ID:    id,
				Name:  name,
				Type:  ft.Type,
				Level: appConfig.DefaultLevel,
				Muted: appConfig.DefaultMuted,
			}
		}
	}

	// Populate store from appConfig.MasterFaders
	for _, mf := range appConfig.MasterFaders {
		log.Printf("Loading master fader: %s (ID: %s)", mf.Name, mf.ID)
		id := mf.ID
		store[id] = &Fader{
			ID:    id,
			Name:  mf.Name,
			Type:  mf.Type,
			Level: appConfig.DefaultLevel,
			Muted: appConfig.DefaultMuted,
		}
	}
	log.Printf("Fader store initialized with %d faders from config.", len(store))
	return store, nil
}
