package main // Stays in package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"m32osc_controller/core" // Import the new core package
)

// StartTCPServer initializes and starts the TCP server on the given listenAddress.
// It now accepts a *core.AppServer instance.
func StartTCPServer(listenAddress string, app *core.AppServer) {
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Fatalf("Failed to start TCP server on %s: %v", listenAddress, err)
		return
	}
	defer listener.Close()
	log.Printf("TCP server listening on %s", listenAddress)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting TCP connection: %v", err)
			continue
		}
		go handleTCPConnection(conn, app)
	}
}

// handleTCPConnection manages a single client TCP connection.
// It now accepts a *core.AppServer instance.
func handleTCPConnection(conn net.Conn, app *core.AppServer) {
	clientAddr := conn.RemoteAddr().String()
	log.Printf("TCP client connected: %s", clientAddr)

	// Create and subscribe a channel for fader updates using app.Events
	subChan := make(core.Subscriber, 10) // Use core.Subscriber
	app.Events.Subscribe(subChan)
	log.Printf("TCP client %s subscribed to FaderEvents.", clientAddr)

	defer func() {
		log.Printf("TCP client %s disconnecting, unsubscribing from FaderEvents", clientAddr)
		app.Events.Unsubscribe(subChan)
		conn.Close()
		log.Printf("TCP client %s disconnected", clientAddr)
	}()

	// Goroutine to listen for updates on subChan and send to client
	go func() {
		for update := range subChan { // update is of type core.FaderUpdate (*core.Fader)
			if update != nil {
				updateMsg := fmt.Sprintf("UPDATE: %s", formatFader(update)) // formatFader takes *core.Fader
				_, err := fmt.Fprintln(conn, updateMsg)
				if err != nil {
					log.Printf("TCP Update: Error writing update to client %s: %v. Update goroutine exiting.", clientAddr, err)
					return
				}
				log.Printf("TCP Update: Sent update for fader %s to client %s", update.ID, clientAddr)
			}
		}
		log.Printf("TCP Update: Subscriber channel closed for client %s. Update goroutine exiting.", clientAddr)
	}()

	fmt.Fprintln(conn, "Welcome to M32 Fader Control TCP Server!")
	fmt.Fprintln(conn, "Subscribed to real-time fader updates.")
	fmt.Fprintln(conn, "Available commands: LIST, GET <id>, SET <id> LEVEL <val>, SET <id> MUTE <ON|OFF>, QUIT")

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		commandLine := strings.TrimSpace(scanner.Text())
		if commandLine == "" {
			continue
		}
		log.Printf("TCP CMD from %s: %s", clientAddr, commandLine)
		parts := strings.Fields(commandLine)
		if len(parts) == 0 {
			continue
		}
		command := strings.ToUpper(parts[0])
		switch command {
		case "LIST":
			handleListCommand(conn, app)
		case "GET":
			if len(parts) < 2 {
				fmt.Fprintln(conn, "ERROR: Missing fader ID for GET command. Usage: GET <fader_id>")
				continue
			}
			faderID := strings.ToUpper(parts[1])
			handleGetCommand(conn, faderID, app)
		case "SET":
			handleSetCommand(conn, parts, app)
		case "QUIT", "EXIT":
			fmt.Fprintln(conn, "Goodbye!")
			return
		default:
			fmt.Fprintf(conn, "ERROR: Unknown command '%s'\n", parts[0])
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err := scanner.Err(); err != nil {
		if err != io.EOF && !strings.Contains(err.Error(), "use of closed network connection") {
			log.Printf("Error reading commands from TCP client %s: %v", clientAddr, err)
		}
	}
}

// formatFader now takes *core.Fader
func formatFader(fader *core.Fader) string {
	return fmt.Sprintf("ID: %s, Name: \"%s\", Type: %s, Level: %.1f dB, Muted: %t",
		fader.ID, fader.Name, fader.Type, fader.Level, fader.Muted)
}

// handleListCommand now accepts a *core.AppServer instance.
func handleListCommand(conn net.Conn, app *core.AppServer) {
	if len(app.Store) == 0 {
		fmt.Fprintln(conn, "No faders available.")
		return
	}
	faders := make([]*core.Fader, 0, len(app.Store)) // Slice of *core.Fader
	for _, fader := range app.Store {
		faders = append(faders, fader)
	}
	typeOrder := map[string]int{
		"Channel": 1, "Bus": 2, "Matrix": 3, "DCA": 4, "Master": 5,
	}
	sort.Slice(faders, func(i, j int) bool {
		orderI, okI := typeOrder[faders[i].Type]
		orderJ, okJ := typeOrder[faders[j].Type]
		if okI && okJ {
			if orderI != orderJ {
				return orderI < orderJ
			}
		} else if okI {
			return true
		} else if okJ {
			return false
		}
		return faders[i].ID < faders[j].ID
	})
	fmt.Fprintln(conn, "--- Fader List (Sorted) ---")
	for _, fader := range faders {
		fmt.Fprintln(conn, formatFader(fader))
	}
	fmt.Fprintln(conn, "--- End of List ---")
}

// handleGetCommand now accepts a *core.AppServer instance.
func handleGetCommand(conn net.Conn, faderID string, app *core.AppServer) {
	fader, ok := app.Store[faderID] // app.Store is map[string]*core.Fader
	if !ok {
		fmt.Fprintf(conn, "ERROR: Fader with ID '%s' not found\n", faderID)
		return
	}
	fmt.Fprintln(conn, formatFader(fader))
}

// handleSetCommand now accepts a *core.AppServer instance.
func handleSetCommand(conn net.Conn, parts []string, app *core.AppServer) {
	if len(parts) < 4 {
		fmt.Fprintln(conn, "ERROR: Invalid SET command format. Usage: SET <id> (LEVEL <value> | MUTE <ON|OFF>)")
		return
	}
	faderID := strings.ToUpper(parts[1])
	property := strings.ToUpper(parts[2])
	valueStr := parts[3]
	fader, ok := app.Store[faderID] // app.Store is map[string]*core.Fader
	if !ok {
		fmt.Fprintf(conn, "ERROR: Fader with ID '%s' not found for SET command\n", faderID)
		return
	}
	switch property {
	case "LEVEL":
		level, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			fmt.Fprintf(conn, "ERROR: Invalid level value '%s'. Must be a number.\n", valueStr)
			return
		}
		if level < -90.0 || level > 10.0 {
			fmt.Fprintf(conn, "ERROR: Level value '%s' out of typical range (-90 to +10 dB).\n", valueStr)
			return
		}
		fader.Level = level
		log.Printf("TCP SET: Fader %s level set to %.1f dB by client %s", faderID, level, conn.RemoteAddr().String())
		app.Events.Publish(fader) // app.Events is *core.Publisher, Publish takes *core.Fader
		fmt.Fprintf(conn, "OK: Fader %s level set to %.1f dB\n", faderID, level)
		fmt.Fprintln(conn, formatFader(fader))
	case "MUTE":
		var newMuteState bool
		switch strings.ToUpper(valueStr) {
		case "ON":
			newMuteState = true
		case "OFF":
			newMuteState = false
		default:
			fmt.Fprintf(conn, "ERROR: Invalid MUTE value '%s'. Must be ON or OFF.\n", valueStr)
			return
		}
		fader.Muted = newMuteState
		log.Printf("TCP SET: Fader %s mute set to %t by client %s", faderID, newMuteState, conn.RemoteAddr().String())
		app.Events.Publish(fader) // app.Events is *core.Publisher, Publish takes *core.Fader
		fmt.Fprintf(conn, "OK: Fader %s mute set to %t\n", faderID, newMuteState)
		fmt.Fprintln(conn, formatFader(fader))
	default:
		fmt.Fprintf(conn, "ERROR: Unknown property '%s' for SET command. Use LEVEL or MUTE.\n", property)
	}
}
