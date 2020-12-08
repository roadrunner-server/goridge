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
func NewPipeRelay(in io.ReadCloser, out io.WriteCloser) Relay {
	// init lookup table for the PipeRelay
	initLookupTable()
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
	header := ReadHeader(hb)
	// we have options
	if header.readHL() > 2 {
		// we should read the options
		optsLen := (header.readHL() - 2) * WORD
		opts := make([]byte, optsLen)
		_, err = rl.in.Read(opts)
		if err != nil {
			return errors.E(op, err)
		}
		header.AppendOptions(opts)
	}

	// verify header CRC
	if header.VerifyCRC() == false {
		return errors.E(op, errors.Str("CRC verification failed"))
	}

	// read the read payload
	pb := make([]byte, header.ReadPayloadLen())
	_, err = rl.in.Read(pb)
	if err != nil {
		return errors.E(op, err)
	}

	*frame = Frame{
		payload: pb,
		header:  header.header,
	}

	return nil
}

// Close the connection. Pipes are closed automatically with the underlying process.
func (rl *PipeRelay) Close() error {
	return nil
}
