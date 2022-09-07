package pipe

import (
	"io"

	"github.com/roadrunner-server/errors"
	"github.com/roadrunner-server/goridge/v3/internal"
	"github.com/roadrunner-server/goridge/v3/pkg/frame"
)

// Relay ... PipeRelay communicate with underlying process using standard streams (STDIN, STDOUT). Attention, use TCP alternative for
// Windows as more reliable option. This relay closes automatically with the process.
type Relay struct {
	in  io.ReadCloser
	out io.WriteCloser
}

// NewPipeRelay creates new pipe based data relay.
func NewPipeRelay(in io.ReadCloser, out io.WriteCloser) *Relay {
	internal.Preallocate()
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
	if frame == nil {
		return errors.Str("nil frame")
	}
	return internal.ReceiveFrame(rl.in, frame)
}

// Close the connection
func (rl *Relay) Close() error {
	_ = rl.out.Close()
	_ = rl.in.Close()
	return nil
}
