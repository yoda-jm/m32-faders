package core // Changed package name

import (
	"fmt"
	// "sync" // Not needed yet
)

// AppServer holds the central application state and components.
// Fader, Publisher are now types defined within this 'core' package.
type AppServer struct {
	Store  map[string]*Fader // Fader store
	Events *Publisher        // Event publisher
}

// NewAppServer creates and returns a new AppServer instance.
// It will initialize the store and publisher.
func NewAppServer(configPath string) (*AppServer, error) {
	store, err := NewFaderStore(configPath) // NewFaderStore is in this package
	if err != nil {
		return nil, fmt.Errorf("failed to create fader store: %w", err)
	}

	publisher := NewPublisher() // NewPublisher is in this package

	return &AppServer{
		Store:  store,
		Events: publisher,
	}, nil
}
