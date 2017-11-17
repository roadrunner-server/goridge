package goridge

const (
	// BufferSize defines max amount of bytes to read from connection at once.
	BufferSize = 655336
)

// Relay provide IPC over signed payloads.
type Relay interface {
	// Send signed (prefixed) data to PHP process.
	Send(data []byte, flags byte) (err error)

	// Receive data from the underlying process and returns associated prefix or error.
	Receive() (data []byte, p Prefix, err error)

	// Close the connection.
	Close() error
}
