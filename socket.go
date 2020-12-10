package goridge

import (
	"io"

	"github.com/spiral/errors"
)

// SocketRelay communicates with underlying process using sockets (TPC or Unix).
type SocketRelay struct {
	rwc io.ReadWriteCloser
}

// NewSocketRelay creates new socket based data relay.
func NewSocketRelay(rwc io.ReadWriteCloser) Relay {
	return &SocketRelay{rwc: rwc}
}

// Send signed (prefixed) data to PHP process.
func (rl *SocketRelay) Send(frame *Frame) error {
	const op = errors.Op("pipes frame send")
	_, err := rl.rwc.Write(frame.Bytes())
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// Receive data from the underlying process and returns associated prefix or error.
func (rl *SocketRelay) Receive(frame *Frame) error {
	const op = errors.Op("pipes frame receive")
	// header bytes
	hb := make([]byte, 8, 8)
	_, err := rl.rwc.Read(hb)
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
		_, err = rl.rwc.Read(opts)
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
	_, err = io.ReadFull(rl.rwc, pb)
	if err != nil {
		return errors.E(op, err)
	}

	*frame = Frame{
		payload: pb,
		header:  header.header,
	}

	return nil
}

// Close the connection.
func (rl *SocketRelay) Close() error {
	return rl.rwc.Close()
}
