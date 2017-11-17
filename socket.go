package goridge

import (
	"io"
	"sync"
)

// SocketRelay communicates with underlying process using sockets (TPC or Unix).
type SocketRelay struct {
	// How many bytes to write/read at once.
	BufferSize uint64

	muw sync.Mutex // concurrent write
	mur sync.Mutex // concurrent read
	rwc io.ReadWriteCloser
}

// NewSocketRelay creates new socket based data relay.
func NewSocketRelay(rwc io.ReadWriteCloser) *SocketRelay {
	return &SocketRelay{BufferSize: BufferSize, rwc: rwc}
}

// Send signed (prefixed) data to PHP process.
func (rl *SocketRelay) Send(data []byte, flags byte) (err error) {
	rl.muw.Lock()
	defer rl.muw.Unlock()

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
	rl.mur.Lock()
	defer rl.mur.Unlock()

	if _, err := rl.rwc.Read(p[:]); err != nil {
		return nil, p, err
	}

	if !p.HasPayload() {
		return nil, p, nil
	}

	data = make([]byte, 0, p.Size())
	leftBytes := p.Size()
	buffer := make([]byte, min(uint64(cap(data)), rl.BufferSize))

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
	rl.muw.Lock()
	rl.mur.Lock()
	defer rl.muw.Unlock()
	defer rl.mur.Unlock()

	return rl.rwc.Close()
}
