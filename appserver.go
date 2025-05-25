package main

import (
	"fmt"
	// "sync" // Not needed yet
)

// AppServer holds the central application state and components.
type AppServer struct {
	Store     map[string]*Fader // Fader store
	Events    *Publisher        // Event publisher
	// Potentially other shared resources like config could go here
}

// NewAppServer creates and returns a new AppServer instance.
// It will initialize the store and publisher.
func NewAppServer(configPath string) (*AppServer, error) {
	store, err := NewFaderStore(configPath) // NewFaderStore will be defined in store.go
	if err != nil {
		return nil, fmt.Errorf("failed to create fader store: %w", err)
	}

	publisher := NewPublisher() // NewPublisher is defined in pubsub.go

	return &AppServer{
		Store:  store,
		Events: publisher,
	}, nil
}
