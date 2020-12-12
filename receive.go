package goridge

import (
	"io"

	"github.com/spiral/errors"
)

func receiveFrame(relay io.Reader, frame *Frame) error {
	const op = errors.Op("pipes frame receive")
	// header bytes
	hb := make([]byte, 12)
	_, err := io.ReadFull(relay, hb)
	if err != nil {
		return errors.E(op, err)
	}

	// Read frame header
	header := ReadHeader(hb)
	// we have options
	if header.readHL() > 3 {
		// we should read the options
		optsLen := (header.readHL() - 3) * WORD
		opts := make([]byte, optsLen)
		_, err := io.ReadFull(relay, opts)
		if err != nil {
			return errors.E(op, err)
		}
		header.AppendOptions(opts)
	}

	// verify header CRC
	if !header.VerifyCRC() {
		return errors.E(op, errors.Str("CRC verification failed"))
	}

	// read the read payload
	pb := make([]byte, header.ReadPayloadLen())
	_, err = io.ReadFull(relay, pb)
	if err != nil {
		return errors.E(op, err)
	}

	*frame = Frame{
		payload: pb,
		header:  header.header,
	}

	return nil
}
