package goridge

import (
	"io"

	"github.com/spiral/errors"
)

// SocketRelay communicates with underlying process using sockets (TPC or Unix).
type SocketRelay struct {
	rwc io.ReadWriteCloser
}

// NewSocketRelay creates new socket based data relay.
func NewSocketRelay(rwc io.ReadWriteCloser) Relay {
	return &SocketRelay{rwc: rwc}
}

// Send signed (prefixed) data to PHP process.
func (rl *SocketRelay) Send(frame *Frame) error {
	const op = errors.Op("pipes frame send")
	_, err := rl.rwc.Write(frame.Bytes())
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// Receive data from the underlying process and returns associated prefix or error.
func (rl *SocketRelay) Receive(frame *Frame) error {
	return receiveFrame(rl.rwc, frame)
}

// Close the connection.
func (rl *SocketRelay) Close() error {
	return rl.rwc.Close()
}
