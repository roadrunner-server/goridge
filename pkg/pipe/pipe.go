package pipe

import (
	"io"

	"github.com/spiral/errors"
	"github.com/spiral/goridge/v3/internal"
	"github.com/spiral/goridge/v3/pkg/frame"
)

// PipeRelay communicate with underlying process using standard streams (STDIN, STDOUT). Attention, use TCP alternative for
// Windows as more reliable option. This relay closes automatically with the process.
type Relay struct {
	in  io.ReadCloser
	out io.WriteCloser
}

// NewPipeRelay creates new pipe based data relay.
func NewPipeRelay(in io.ReadCloser, out io.WriteCloser) *Relay {
	return &Relay{in: in, out: out}
}

// Send signed (prefixed) data to underlying process.
func (rl *Relay) Send(frame *frame.Frame) error {
	const op = errors.Op("pipes frame send")
	_, err := rl.out.Write(frame.Bytes())
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (rl *Relay) Receive(frame *frame.Frame) error {
	return internal.ReceiveFrame(rl.in, frame)
}

// Close the connection. Pipes are closed automatically with the underlying process.
func (rl *Relay) Close() error {
	return nil
}
