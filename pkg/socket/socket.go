package socket

import (
	"io"

	"github.com/spiral/errors"
	"github.com/spiral/goridge/v3/internal"
	"github.com/spiral/goridge/v3/pkg/frame"
)

// Relay communicates with underlying process using sockets (TPC or Unix).
type Relay struct {
	rwc io.ReadWriteCloser
}

// NewSocketRelay creates new socket based data relay.
func NewSocketRelay(rwc io.ReadWriteCloser) *Relay {
	return &Relay{rwc: rwc}
}

// Send signed (prefixed) data to PHP process.
func (rl *Relay) Send(frame *frame.Frame) error {
	const op = errors.Op("pipes frame send")
	_, err := rl.rwc.Write(frame.Bytes())
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// Receive data from the underlying process and returns associated prefix or error.
func (rl *Relay) Receive(frame *frame.Frame) error {
	return internal.ReceiveFrame(rl.rwc, frame)
}

// Close the connection.
func (rl *Relay) Close() error {
	return rl.rwc.Close()
}
