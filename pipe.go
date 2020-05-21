package goridge

import (
	"errors"
	"io"
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

// Receive data from the underlying process and returns associated prefix or error.
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

	// Here can be 3 cases
	// n > 0, then we make a syscall to read all data from the Relay
	// n == 0, no need to make an extra call, because we do not expect any data from the Relay
	// n < 0, impossible, since the p.Size() is uint64
	switch n := p.Size(); {
	// LIKELY
	case n > 0:
		data = make([]byte, n)
		n, err := rl.in.Read(data)
		if err != nil {
			return nil, p, err
		}
		// ensure, that we read all the provided data
		if uint64(n) == p.Size() {
			return data, p, nil
		}
		return nil, p, errors.New("read only part of the data from the pipe relay")

		// POSSIBLE
	case n == 0:
		// return valid prefix, w/o data and w/o error
		return nil, p, nil
		// IMPOSSIBLE
	default:
		return nil, p, errors.New("unexpected case in the pipes relay")
	}
}

// Close the connection. Pipes are closed automatically with the underlying process.
func (rl *PipeRelay) Close() error {
	return nil
}
