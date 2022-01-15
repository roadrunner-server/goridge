package relay

import "github.com/roadrunner-server/goridge/v3/pkg/frame"

// Relay provide IPC over signed payloads.
type Relay interface {
	// Send signed (prefixed) data to PHP process.
	Send(frame *frame.Frame) error

	// Receive data from the underlying process and returns associated prefix or error.
	Receive(frame *frame.Frame) error

	// Close the connection.
	Close() error
}
