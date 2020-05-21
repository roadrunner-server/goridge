package goridge

import (
	"errors"
	"io"
)

// SocketRelay communicates with underlying process using sockets (TPC or Unix).
type SocketRelay struct {
	rwc io.ReadWriteCloser
}

// NewSocketRelay creates new socket based data relay.
func NewSocketRelay(rwc io.ReadWriteCloser) *SocketRelay {
	return &SocketRelay{rwc: rwc}
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

	maxAlloc := getAllocSize()

	// Here can be 4 cases
	// n > 0 and n < maxAlloc, then we make a syscall to read all data from the Relay
	// n == 0, no need to make an extra call, because we do not expect any data from the Relay
	// n > maxAlloc - error, cannot allocate such bit slice
	// n < 0, impossible, since the p.Size() is uint64
	switch n := p.Size(); {
	// LIKELY
	case n > 0 && uint(n) < maxAlloc:
		data = make([]byte, n)
		n, err := rl.rwc.Read(data)
		if err != nil {
			return nil, p, err
		}
		// ensure, that we read all the provided data
		if uint64(n) == p.Size() {
			return data, p, nil
		}
		return nil, p, errors.New("read only part of the data from the socket relay")

		// POSSIBLE
	case uint(n) >= maxAlloc:
		return nil, p, errors.New("cannot allocate more then 17.1 Gb on x64 or 2.14 on x86 systems")
		// POSSIBLE
	case n == 0:
		// return valid prefix, w/o data and w/o error
		return nil, p, nil
		// IMPOSSIBLE
	default:
		return nil, p, errors.New("unexpected case in the socket relay")
	}
}

// Close the connection.
func (rl *SocketRelay) Close() error {
	return rl.rwc.Close()
}
