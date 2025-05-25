package main

import "fmt"

// MasterFaderStore holds all the faders available on the console.
// The key is the Fader.ID.
var MasterFaderStore map[string]*Fader

func init() {
	MasterFaderStore = make(map[string]*Fader)

	// 32 Input Channels
	for i := 1; i <= 32; i++ {
		id := fmt.Sprintf("CH%02d", i)
		MasterFaderStore[id] = &Fader{
			ID:    id,
			Name:  fmt.Sprintf("Channel %d", i),
			Type:  "Channel",
			Level: 0.0,
			Muted: false,
		}
	}

	// 16 Buses
	for i := 1; i <= 16; i++ {
		id := fmt.Sprintf("BUS%02d", i)
		MasterFaderStore[id] = &Fader{
			ID:    id,
			Name:  fmt.Sprintf("Bus %d", i),
			Type:  "Bus",
			Level: 0.0,
			Muted: false,
		}
	}

	// 8 DCAs
	for i := 1; i <= 8; i++ {
		id := fmt.Sprintf("DCA%02d", i)
		MasterFaderStore[id] = &Fader{
			ID:    id,
			Name:  fmt.Sprintf("DCA %d", i),
			Type:  "DCA",
			Level: 0.0,
			Muted: false,
		}
	}

	// 6 Matrices
	for i := 1; i <= 6; i++ {
		id := fmt.Sprintf("MTX%02d", i)
		MasterFaderStore[id] = &Fader{
			ID:    id,
			Name:  fmt.Sprintf("Matrix %d", i),
			Type:  "Matrix",
			Level: 0.0,
			Muted: false,
		}
	}

	// 1 Main Master Fader
	MasterFaderStore["MAINLR"] = &Fader{
		ID:    "MAINLR",
		Name:  "Main LR",
		Type:  "Master",
		Level: 0.0,
		Muted: false,
	}
}
