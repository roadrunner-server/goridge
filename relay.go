package goridge

const (
	Version        = 0
	PayloadControl = 1
	PayloadError   = 2
)

// Frame defines new user level package format.
type Frame struct {
	// Flags are domain specific byte.
	Flags byte

	// Options in free format, max length 60 bytes.
	Options []byte

	// Payload, max length 4.2GB.
	Payload []byte
}

// HasFlag returns true if prefix has given flag.
func (f Frame) HasFlag(flag byte) bool {
	return f.Flags&flag == flag
}

// Relay provide IPC over signed payloads.
type Relay interface {
	// Send signed (prefixed) data to PHP process.
	Send(Frame) error

	// Receive data from the underlying process and returns associated prefix or error.
	Receive() (Frame, error)

	// Close the connection.
	Close() error
}
