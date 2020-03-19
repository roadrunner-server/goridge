package goridge

import (
	"errors"
	"io"
)

// SocketRelay communicates with underlying process using sockets (TPC or Unix).
type SocketRelay struct {
	// How many bytes to write/read at once.
	BufferSize uint64
	rwc io.ReadWriteCloser
}

// NewSocketRelay creates new socket based data relay.
func NewSocketRelay(rwc io.ReadWriteCloser) *SocketRelay {
	return &SocketRelay{BufferSize: BufferSize, rwc: rwc}
}

// Send signed (prefixed) data to PHP process.
func (rl *SocketRelay) Send(data []byte, flags byte) (err error) {
	prefix := NewPrefix().WithFlags(flags).WithSize(uint64(len(data)))
	if _, err := rl.rwc.Write(prefix[:]); err != nil {
		return err
	}

	if _, err := rl.rwc.Write(data); err != nil {
		return err
	}

	return nil
}

// Receive data from the underlying process and returns associated prefix or error.
func (rl *SocketRelay) Receive() (data []byte, p Prefix, err error) {
	defer func() {
		if rErr, ok := recover().(error); ok {
			err = rErr
		}
	}()

	if _, err := rl.rwc.Read(p[:]); err != nil {
		return nil, p, err
	}

	if !p.Valid() {
		return nil, p, errors.New("invalid data found in the buffer (possible echo)")
	}

	if !p.HasPayload() {
		return nil, p, nil
	}

	leftBytes := p.Size()
	data = make([]byte, 0, leftBytes)
	buffer := make([]byte, min(leftBytes, rl.BufferSize))

	for {
		if n, err := rl.rwc.Read(buffer); err == nil {
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

// Close the connection.
func (rl *SocketRelay) Close() error {
	return rl.rwc.Close()
}
