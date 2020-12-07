package goridge

import (
	"io"

	"github.com/spiral/errors"
)

// PipeRelay communicate with underlying process using standard streams (STDIN, STDOUT). Attention, use TCP alternative for
// Windows as more reliable option. This relay closes automatically with the process.
type PipeRelay struct {
	in  io.ReadCloser
	out io.WriteCloser
}

// NewPipeRelay creates new pipe based data relay.
func NewPipeRelay(in io.ReadCloser, out io.WriteCloser) *PipeRelay {
	return &PipeRelay{in: in, out: out}
}

// Send signed (prefixed) data to underlying process.
func (rl *PipeRelay) Send(frame *Frame) error {
	const op = errors.Op("pipes frame send")
	_, err := rl.out.Write(frame.Bytes())
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (rl *PipeRelay) Receive(frame *Frame) error {
	const op = errors.Op("pipes frame receive")
	// header bytes
	hb := make([]byte, 8, 8)
	_, err := rl.in.Read(hb)
	if err != nil {
		return errors.E(op, err)
	}

	// Read frame header
	header := ReadFrame(hb)
	if header.VerifyCRC() == false {
		return errors.E(op, errors.Str("CRC verification failed"))
	}

	// payload bytes
	pb := make([]byte, header.ReadPayloadLen())
	_, err = rl.in.Read(pb)
	if err != nil {
		return errors.E(op, err)
	}

	*frame = *ReadFrame(pb)
	frame.MergeHeader(header)

	return nil
}

// Close the connection. Pipes are closed automatically with the underlying process.
func (rl *PipeRelay) Close() error {
	return nil
}
