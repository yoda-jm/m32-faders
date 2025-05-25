package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

// StartTCPServer initializes and starts the TCP server.
func StartTCPServer(port string) {
	listenAddress := ":" + port
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Fatalf("Failed to start TCP server on port %s: %v", port, err)
		return
	}
	defer listener.Close()
	log.Printf("TCP server listening on port %s", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting TCP connection: %v", err)
			continue // Continue accepting other connections
		}
		// Handle each connection in a new goroutine
		go handleTCPConnection(conn)
	}
}

// handleTCPConnection manages a single client TCP connection.
func handleTCPConnection(conn net.Conn) {
	clientAddr := conn.RemoteAddr().String()
	log.Printf("TCP client connected: %s", clientAddr)
	defer func() {
		conn.Close()
		log.Printf("TCP client disconnected: %s", clientAddr)
	}()

	// Welcome message
	fmt.Fprintln(conn, "Welcome to M32 Fader Control TCP Server!")
	fmt.Fprintln(conn, "Available commands: LIST, GET <id>, SET <id> LEVEL <val>, SET <id> MUTE <ON|OFF>, QUIT")

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		commandLine := strings.TrimSpace(scanner.Text())
		if commandLine == "" {
			continue
		}

		log.Printf("TCP CMD from %s: %s", clientAddr, commandLine)
		parts := strings.Fields(commandLine) // Split by whitespace
		if len(parts) == 0 {
			continue
		}

		command := strings.ToUpper(parts[0])

		switch command {
		case "LIST":
			handleListCommand(conn)
		case "GET":
			if len(parts) < 2 {
				fmt.Fprintln(conn, "ERROR: Missing fader ID for GET command. Usage: GET <fader_id>")
				continue
			}
			faderID := strings.ToUpper(parts[1])
			handleGetCommand(conn, faderID)
		case "SET":
			handleSetCommand(conn, parts)
		case "QUIT", "EXIT":
			fmt.Fprintln(conn, "Goodbye!")
			return // Close connection via defer
		default:
			fmt.Fprintf(conn, "ERROR: Unknown command '%s'\n", parts[0])
		}
		// Add a small delay to prevent tight loop CPU spinning if client sends commands too fast
		// and there are no blocking operations in handlers.
		time.Sleep(50 * time.Millisecond)
	}

	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			log.Printf("Error reading from TCP client %s: %v", clientAddr, err)
		}
	}
}

func formatFader(fader *Fader) string {
	return fmt.Sprintf("ID: %s, Name: \"%s\", Type: %s, Level: %.1f dB, Muted: %t",
		fader.ID, fader.Name, fader.Type, fader.Level, fader.Muted)
}

func handleListCommand(conn net.Conn) {
	// Note: Accessing MasterFaderStore directly. Consider mutex for concurrent access if map structure changes.
	if len(MasterFaderStore) == 0 {
		fmt.Fprintln(conn, "No faders available.")
		return
	}
	fmt.Fprintln(conn, "--- Fader List ---")
	for _, fader := range MasterFaderStore { // Iterate over values (pointers to Fader)
		fmt.Fprintln(conn, formatFader(fader))
	}
	fmt.Fprintln(conn, "--- End of List ---")
}

func handleGetCommand(conn net.Conn, faderID string) {
	fader, ok := MasterFaderStore[faderID]
	if !ok {
		fmt.Fprintf(conn, "ERROR: Fader with ID '%s' not found\n", faderID)
		return
	}
	fmt.Fprintln(conn, formatFader(fader))
}

func handleSetCommand(conn net.Conn, parts []string) {
	// Expected format: SET <fader_id> <PROPERTY> <value>
	// e.g., SET CH01 LEVEL -10.5
	// e.g., SET BUS02 MUTE ON
	if len(parts) < 4 {
		fmt.Fprintln(conn, "ERROR: Invalid SET command format. Usage: SET <id> (LEVEL <value> | MUTE <ON|OFF>)")
		return
	}

	faderID := strings.ToUpper(parts[1])
	property := strings.ToUpper(parts[2])
	valueStr := parts[3]

	fader, ok := MasterFaderStore[faderID]
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
		// Basic validation for level, can be expanded
		if level < -90.0 || level > 10.0 {
			fmt.Fprintf(conn, "ERROR: Level value '%s' out of typical range (-90 to +10 dB).\n", valueStr)
			return
		}
		fader.Level = level // Direct update to the struct field
		log.Printf("TCP SET: Fader %s level set to %.1f dB by client %s", faderID, level, conn.RemoteAddr().String())
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
		fader.Muted = newMuteState // Direct update
		log.Printf("TCP SET: Fader %s mute set to %t by client %s", faderID, newMuteState, conn.RemoteAddr().String())
		fmt.Fprintf(conn, "OK: Fader %s mute set to %t\n", faderID, newMuteState)
		fmt.Fprintln(conn, formatFader(fader))

	default:
		fmt.Fprintf(conn, "ERROR: Unknown property '%s' for SET command. Use LEVEL or MUTE.\n", property)
	}
}
