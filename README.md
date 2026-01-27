# go-guac

Go library for Apache Guacamole protocol implementation.

## Installation

```bash
go get github.com/riete/go-guac
```

## Packages

- `protocol` - Guacamole protocol instructions and handshake
- `tunnel` - Tunnel management between guacd and WebSocket
- `recorder` - Session recording with optional gzip compression

## Quick Start

```go
package main

import (
    "context"
    "net"

    "github.com/gorilla/websocket"
    "gitlab.bangdao-tech.com/kubeease/resource/go-guac/protocol"
    "gitlab.bangdao-tech.com/kubeease/resource/go-guac/recorder"
    "gitlab.bangdao-tech.com/kubeease/resource/go-guac/tunnel"
)

func main() {
    // Connect to guacd
    guacd, _ := net.Dial("tcp", "localhost:4822")
    
    // WebSocket connection (from HTTP upgrade)
    var ws *websocket.Conn
    
    // Create recorder (optional)
    rec := recorder.NewFileRecorder(
        recorder.WithBaseDirectory("/path/to/records"),
        recorder.WithGzipCompress(),
    )
    
    // Create tunnel
    t := tunnel.NewTunnel(guacd, ws,
        tunnel.WithGuacdKeepalive(time.Minute),
        tunnel.WithWsKeepalive(30*time.Second, 2),
        tunnel.WithRecorder(rec),
        tunnel.WithOnConnect(func(connId string) {
            log.Printf("Connected: %s", connId)
        }),
        tunnel.WithOnDisconnect(func(connId string) {
            log.Printf("Disconnected: %s", connId)
        }),
    )
    defer t.Close()
    
    // Handshake configuration
    config := protocol.NewHandshakeConfig(
        map[string]string{
            "hostname": "192.168.1.100",
            "port":     "3389",
            "username": "admin",
            "password": "secret",
        },
        protocol.WithProtocol("rdp"),
        protocol.WithScreenSize(1920, 1080, 96),
    )
    
    // Perform handshake
    if err := t.Handshake(config); err != nil {
        log.Fatal(err)
    }
    
    // Forward data
    ctx := context.Background()
    if err := t.Forward(ctx); err != nil {
        log.Printf("Forward error: %v", err)
    }
}
```

## Protocol Package

### Instructions

```go
// Create instruction
instr := protocol.NewInstruction("select", "rdp")
// Output: 6.select,3.rdp;

// Parse instruction
instr := protocol.Instruction("6.select,3.rdp;")
opcode := instr.Opcode().Value()  // "select"
args := instr.Args()              // []Element

// Check for error instruction
if instr.IsError() {
    err := instr.Error()  // Returns formatted error
}

// Global instructions
protocol.Nop        // nop instruction
protocol.Disconnect // disconnect instruction
```

### Handshake Configuration

```go
config := protocol.NewHandshakeConfig(
    map[string]string{
        "hostname": "192.168.1.100",
        "port":     "3389",
        "username": "admin",
        "password": "secret",
    },
    // Protocol
    protocol.WithProtocol("rdp"),  // rdp, vnc, ssh
    
    // Screen settings
    protocol.WithScreenSize(1920, 1080, 96),
    
    // Audio/Video codecs
    protocol.WithAudioCodecs("audio/L8", "audio/L16"),
    protocol.WithVideoCodecs("video/webm"),
    protocol.WithImageFormats("image/png", "image/jpeg"),
    
    // RDP specific options
    protocol.WithSecurity("nla"),          // nla, tls, rdp, any
    protocol.WithIgnoreCert(true),
    protocol.WithDisableAuth(false),
    protocol.WithEnableDrive(true),
    protocol.WithDrivePath("/shared"),
    protocol.WithConsole(false),
    
    // SSH specific options
    protocol.WithPrivateKey("..."),
    protocol.WithPassphrase("..."),
)
```

### Status Codes

```go
status := protocol.Success           // 0
status := protocol.ClientUnauthorized // 769

// Get string representation
str := status.String()  // "769_CLIENT_UNAUTHORIZED"
```

## Tunnel Package

### Options

```go
tunnel.NewTunnel(guacd, ws,
    // Keepalive settings
    tunnel.WithGuacdKeepalive(time.Minute),      // Send nop to guacd
    tunnel.WithWsKeepalive(30*time.Second, 2),   // Ping/pong with deadline
    
    // Recorder
    tunnel.WithRecorder(recorder),
    
    // Callbacks (chainable, called in order)
    tunnel.WithOnConnect(func(connId string) { }),
    tunnel.WithOnDisconnect(func(connId string) { }),
    tunnel.WithOnReadFromGuacd(func(connId string, data []byte) { }),
    tunnel.WithOnReadFromWs(func(connId string, data []byte) { }),
)
```

### Methods

```go
// Perform handshake
err := t.Handshake(config)

// Get connection ID
connId := t.ConnId()

// Forward data (blocks until context cancelled or error)
err := t.Forward(ctx)

// Close tunnel
t.Close()
```

## Recorder Package

### FileRecorder

```go
// Create recorder
rec := recorder.NewFileRecorder(
    recorder.WithBaseDirectory("/path/to/records"),
    recorder.WithGzipCompress(),  // Enable gzip compression
)

// Record data
rec.Record(connId, data)

// Close recording
rec.Close(connId)

// Replay recording
ctx := context.Background()
ch, err := rec.Replay(ctx, connId)
if err != nil {
    log.Fatal(err)
}
for instruction := range ch {
    fmt.Println(instruction)
}
```

### Integration with Tunnel

```go
rec := recorder.NewFileRecorder(
    recorder.WithBaseDirectory("/records"),
    recorder.WithGzipCompress(),
)

t := tunnel.NewTunnel(guacd, ws,
    tunnel.WithRecorder(rec),  // Automatically records and closes
)
```

## Recorder Interface

Implement custom recorders:

```go
type Recorder interface {
    Record(connId string, data []byte)
    Replay(ctx context.Context, connId string) (chan string, error)
    Close(connId string)
}
```
