package internal

import (
	"io"

	"github.com/spiral/errors"
	"github.com/spiral/goridge/v3/pkg/frame"
)

func ReceiveFrame(relay io.Reader, fr *frame.Frame) error {
	const op = errors.Op("goridge_frame_receive")
	// header bytes
	hb := make([]byte, 12)
	_, err := io.ReadFull(relay, hb)
	if err != nil {
		return errors.E(op, err)
	}

	// Read frame header
	header := frame.ReadHeader(hb)
	// we have options
	if header.ReadHL() > 3 {
		// we should read the options
		optsLen := (header.ReadHL() - 3) * frame.WORD
		opts := make([]byte, optsLen)
		_, err = io.ReadFull(relay, opts)
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

	*fr = *frame.From(header.Header(), pb)

	return nil
}
