package goridge

import (
	"errors"
	"io"
	"math"
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
func (rl *PipeRelay) Send(data []byte, flags byte) (err error) {
	prefix := NewPrefix().WithFlags(flags).WithSize(uint64(len(data)))
	if _, err := rl.out.Write(append(prefix[:], data...)); err != nil {
		return err
	}

	return nil
}

func (rl *PipeRelay) Receive() (data []byte, p Prefix, err error) {
	defer func() {
		if rErr, ok := recover().(error); ok {
			err = rErr
		}
	}()

	if _, err := rl.in.Read(p[:]); err != nil {
		return nil, p, err
	}

	if !p.Valid() {
		return nil, p, errors.New("invalid data found in the buffer (possible echo)")
	}

	if !p.HasPayload() {
		return nil, p, nil
	}

	data = make([]byte, 0, p.Size())
	leftBytes := p.Size()
	buffer := make([]byte, uint(math.Min(float64(cap(data)), float64(BufferSize))))
	for {
		if n, err := rl.in.Read(buffer); err == nil {
			data = append(data, buffer[:n]...)
			leftBytes -= uint64(n)
		} else {
			return nil, p, err
		}

		if leftBytes == 0 {
			break
		}
	}

	return
}

// Close the connection. Pipes are closed automatically with the underlying process.
func (rl *PipeRelay) Close() error {
	return nil
}
