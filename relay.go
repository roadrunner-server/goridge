package goridge

// Relay provide IPC over signed payloads.
type Relay interface {
	// Send signed (prefixed) data to PHP process.
	Send(frame *Frame) error

	// Receive data from the underlying process and returns associated prefix or error.
	Receive(frame *Frame) error

	// Close the connection.
	Close() error
}
