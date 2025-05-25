# M32 Fader Web/TCP Control Server

## Description

A Go application providing a web interface and a TCP command server to control virtual faders, mimicking an M32 audio console's fader controls. It includes real-time updates via WebSockets for the web UI and pushes updates to connected TCP clients. The fader layout (channels, buses, etc.) is configurable via an external JSON file.

## Features

*   **Web UI**: For fader selection and interactive control (level, mute).
*   **Real-time Web Updates**: Fader changes are reflected live in the Web UI via WebSockets.
*   **Double-Click Reset**: Double-clicking a fader slider in the UI resets its level to 0dB.
*   **TCP Command Server**: Allows command-line control of faders.
    *   Supported commands: `LIST`, `GET <id>`, `SET <id> LEVEL <value>`, `SET <id> MUTE <ON|OFF>`.
*   **Real-time TCP Updates**: Fader changes are pushed to all connected TCP clients.
*   **Configurable Fader Layout**: Define channels, buses, matrices, DCAs, and master faders via `config.json`.
*   **Configurable Network Addresses**: HTTP and TCP server addresses, config file path, and OSC target address are configurable via command-line flags.
*   **OSC Client Placeholder**: Basic structure for future Open Sound Control (OSC) client integration.

## Prerequisites

*   **Go**: Version 1.22.2 or newer (as per `go.mod`).

## Setup and Configuration

### 1. Build

From the project root directory, build the application:

```bash
go build .
```

This will produce an executable (e.g., `m32osc_controller` on Linux/macOS or `m32osc_controller.exe` on Windows).

### 2. Configuration (`config.json`)

This file defines the layout of faders in the application. It should be present in the same directory as the executable or its path specified via the `-config` flag.

**Structure:**

*   `faderTypes` (array): Defines groups of faders. Each object has:
    *   `type` (string): Category name (e.g., "Channel", "Bus").
    *   `count` (integer): Number of faders in this group.
    *   `idPrefix` (string): Prefix for generating fader IDs (e.g., "CH", "BUS").
    *   `namePrefix` (string): Prefix for generating fader names (e.g., "Channel ", "Bus ").
*   `masterFaders` (array): Defines individual, uniquely named faders. Each object has:
    *   `type` (string): Category name (e.g., "Master").
    *   `id` (string): Unique ID for the fader (e.g., "MAINLR").
    *   `name` (string): User-friendly name (e.g., "Main LR").
*   `defaultLevel` (float): Initial level for all faders (e.g., `0.0`).
*   `defaultMuted` (boolean): Initial mute state for all faders (e.g., `false`).

**Example `config.json`:**

```json
{
  "faderTypes": [
    {
      "type": "Channel",
      "count": 32,
      "idPrefix": "CH",
      "namePrefix": "Channel "
    },
    {
      "type": "Bus",
      "count": 16,
      "idPrefix": "BUS",
      "namePrefix": "Bus "
    }
    // ... more types
  ],
  "masterFaders": [
    {
      "type": "Master",
      "id": "MAINLR",
      "name": "Main LR"
    }
  ],
  "defaultLevel": 0.0,
  "defaultMuted": false
}
```

### 3. Command-Line Flags

The application supports the following command-line flags:

*   `-http-addr`: HTTP server address. Default: `localhost:8080`. (For Web UI, API, and WebSockets).
*   `-tcp-addr`: TCP command server address. Default: `localhost:8001`.
*   `-config`: Path to the `config.json` file. Default: `config.json`.
*   `-osc-addr`: Target M32 OSC server address (for placeholder OSC client). Default: `127.0.0.1:10023`.

## Running the Application

*   **For development (from project root):**
    ```bash
    go run .
    ```
    You can also use flags with `go run`:
    ```bash
    go run . -http-addr=":8088" -tcp-addr=":9009"
    ```

*   **After building (using the compiled executable):**
    Let the executable be `m32osc_controller`.
    ```bash
    ./m32osc_controller
    ```
    With custom flags:
    ```bash
    ./m32osc_controller -http-addr=":80" -tcp-addr=":9000" -config="/path/to/your/custom_config.json"
    ```

## Usage

### Web UI

1.  **Access**: Open your web browser and navigate to the HTTP address specified by `-http-addr` (e.g., `http://localhost:8080` if using defaults).
2.  **Fader Selection**:
    *   Use the "Fader Type" dropdown to select a category of faders (e.g., "Channel", "Bus", or "All").
    *   Use the "Fader Item/Name" dropdown to select a specific fader.
3.  **Fader Control**:
    *   **Slider**: Drag the slider to change the fader's level. The dB value is displayed next to it.
    *   **Mute Button**: Click to toggle the mute state of the selected fader. The button text and color will change accordingly ("Mute" in red, "Unmute" in green).
    *   **Double-Click Reset**: Double-click directly on the fader slider to reset its level to 0 dB.
4.  **Real-time Updates**: Changes made to a fader (either via the Web UI, TCP server, or potentially other sources in the future) will be reflected live in the Web UI if that fader is currently selected.

### TCP Server

1.  **Connect**: Use a TCP client like `netcat` (nc) or Telnet to connect to the TCP server address.
    ```bash
    nc localhost 8001 
    ```
    (Replace `localhost` and `8001` if using different host/port from flags).

2.  **Commands**:
    *   `LIST`: Lists all available faders, sorted by type and then by ID.
        ```
        LIST
        ```
    *   `GET <fader_id>`: Shows the current state of a specific fader.
        ```
        GET CH01
        GET MAINLR
        ```
    *   `SET <fader_id> LEVEL <value>`: Sets the level of a specific fader. The level is a float (e.g., -10.5, 0, 5.0).
        ```
        SET CH01 LEVEL -10.5
        SET BUS02 LEVEL 0.0
        ```
    *   `SET <fader_id> MUTE <ON|OFF>`: Sets the mute state of a specific fader.
        ```
        SET CH01 MUTE ON
        SET CH01 MUTE OFF
        ```
    *   `QUIT` or `EXIT`: Disconnects the TCP client.

3.  **Real-time Updates**: If a fader's state is changed by any means (Web UI, another TCP client, etc.), an `UPDATE: ...` message with the new fader state will be pushed to all connected TCP clients.

## Project Structure

*   `main.go`: Main application entry point, server initialization, flag parsing.
*   `core/`: Contains core application logic and data structures.
    *   `appserver.go`: Defines `AppServer` struct, centralizing application state.
    *   `types.go`: Defines the `Fader` struct.
    *   `pubsub.go`: Implements the publish/subscribe system for fader events.
    *   `store.go`: Handles loading fader definitions from `config.json`.
*   `webserver/`: Encapsulates all HTTP server functionality.
    *   `server.go`: Initializes the HTTP router, middleware, and starts the HTTP server.
    *   `http_handlers.go`: Contains HTTP API route handlers.
    *   `websocket_handler.go`: Contains WebSocket connection handler and logic.
*   `tcp_server.go`: Implements the TCP command server logic.
*   `osc.go`: Placeholder for future OSC client functionality.
*   `static/`: Contains all static files for the Web UI (HTML, CSS, JavaScript).
*   `config.json`: Default configuration file for fader layout.

## Future Enhancements / TODOs

*   **Full OSC Client**: Implement complete OSC client functionality to send and receive updates from a real Behringer M32/X32 console.
*   **State Persistence**: Persist fader states (e.g., to a file or a simple database) so they are retained across application restarts.
*   **Enhanced UI Feedback**: Add more sophisticated error handling and user feedback in the Web UI (e.g., visual indication of connection status, error messages for failed operations).
*   **OSC Address Mapping**: Allow mapping of internal fader IDs/types to specific OSC addresses in `config.json`.
*   **Security**: Add considerations for security if exposing the application outside a local network (e.g., authentication, HTTPS).
*   **Testing**: Add unit and integration tests.
