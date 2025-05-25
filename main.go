package main

import (
	"flag"
	"log"

	"m32osc_controller/core"     // Import the new core package
	"m32osc_controller/webserver" // Import the webserver package
)

// Types like Fader, AppServer, Publisher, Subscriber are now defined in core package.

func main() {
	// Declare variables for command-line flags
	var httpAddr string
	var tcpAddr string
	var configPath string
	var oscAddr string

	// Define command-line flags using flag.StringVar
	flag.StringVar(&httpAddr, "http-addr", "localhost:8080", "HTTP server address (e.g., localhost:8080 or :8080)")
	flag.StringVar(&tcpAddr, "tcp-addr", "localhost:8001", "TCP server address (e.g., localhost:8001 or :8001)")
	flag.StringVar(&configPath, "config", "config.json", "Path to the configuration file")
	flag.StringVar(&oscAddr, "osc-addr", "127.0.0.1:10023", "Target M32 OSC server address")
	flag.Parse() // Parse the command-line flags

	// Create the AppServer instance using core.NewAppServer
	// Use flag variables directly (no dereferencing '*')
	app, err := core.NewAppServer(configPath) // NewAppServer returns (*core.AppServer, error)
	if err != nil {
		log.Fatalf("Failed to initialize AppServer: %v", err)
	}

	// Start the TCP server in a goroutine using the address from the flag
	// Use flag variables directly
	go StartTCPServer(tcpAddr, app)

	// Start the OSC client placeholder
	// Use flag variables directly
	go StartOSCClient(oscAddr, app)

	// Start the HTTP server (this is a blocking call)
	// Use flag variables directly
	log.Printf("Starting HTTP server on %s via webserver package", httpAddr)
	webserver.Start(httpAddr, app) // This will block and log if ListenAndServe fails
}
