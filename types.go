package main

// Fader represents a single fader on the M32 console.
type Fader struct {
	ID     string  // Unique identifier (e.g., "CH01", "BUS01", "DCA1", "MAINLR")
	Name   string  // User-friendly name (e.g., "Kick Drum", "Vocal Bus", "DCA Group 1", "Main Stereo Out")
	Type   string  // Category of the fader (e.g., "Channel", "Bus", "DCA", "Matrix", "Master")
	Level  float64 // Fader level, typically from -90.0 to +10.0 (dB)
	Muted  bool    // Mute status
}
