package main // Stays in package main

import (
	"log"

	"m32osc_controller/core" // Import the new core package
)

// StartOSCClient initializes the placeholder OSC client.
// It listens for fader updates and logs what OSC messages it would send.
func StartOSCClient(m32Address string, app *core.AppServer) {
	log.Printf("OSC Client: Initializing for target M32 address: %s", m32Address)

	subChan := make(core.Subscriber, 10) // Buffered channel
	app.Events.Subscribe(subChan)
	log.Printf("OSC Client: Subscribed to FaderEvents for %s", m32Address)

	// Ensure unsubscription when this function's goroutine context (if any) ends,
	// or if StartOSCClient were to block, when it returns.
	// Since StartOSCClient launches a goroutine and returns, this defer might not be
	// strictly necessary for the subChan lifecycle if the app closes,
	// but it's good practice if StartOSCClient had more complex blocking logic.
	// The main cleanup for subChan is handled by the goroutine itself if it exits.
	// However, for clarity and safety, if the app itself is shutting down,
	// the publisher should ideally close all subscriber channels.
	// For now, we are focusing on the client's lifecycle.
	// The goroutine below will run until subChan is closed by app.Events.Unsubscribe.
	// If StartOSCClient's goroutine (from main.go) itself is terminated, this defer won't run.
	// The primary unsubscription point for this specific subChan is when the app itself closes or
	// if we were to add specific logic here to stop this OSC client instance.
	// For this placeholder, direct deferral of Unsubscribe is not strictly what we want if StartOSCClient returns immediately.
	// The listener goroutine handles its lifecycle based on subChan.

	go func() {
		// This defer is important for when the subChan is closed.
		defer func() {
			log.Printf("OSC Client: Fader update channel closed for %s. Unsubscribing and goroutine exiting.", m32Address)
			app.Events.Unsubscribe(subChan)
		}()

		log.Printf("OSC Client: Listening for fader updates to send to %s", m32Address)
		for faderUpdate := range subChan { // faderUpdate is *core.Fader
			if faderUpdate != nil {
				sendOSCUpdate(faderUpdate, m32Address)
			}
		}
		// This part of the code (after the loop) will be reached when subChan is closed.
	}()

	log.Println("OSC Client: Placeholder started. Listening for fader updates.")
}

// sendOSCUpdate is a placeholder for sending actual OSC messages.
func sendOSCUpdate(fader *core.Fader, m32Address string) {
	// In a real implementation, this would construct and send an OSC message.
	// For now, it just logs the intent.
	log.Printf("OSC: [Placeholder] Would send update for Fader ID '%s' (Type: %s, Level: %.2f, Muted: %t) to %s",
		fader.ID, fader.Type, fader.Level, fader.Muted, m32Address)
	// Example OSC messages might be:
	// /ch/01/mix/fader [float32: level_0_to_1]
	// /ch/01/mix/on [int32: 0_or_1] (for mute status)
}
