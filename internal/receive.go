package internal

import (
	"io"

	"github.com/spiral/errors"
	"github.com/spiral/goridge/v3/pkg/frame"
)

func ReceiveFrame(relay io.Reader, fr *frame.Frame) error {
	const op = errors.Op("goridge_frame_receive")

	if fr == nil {
		return errors.E(op, errors.Str("nil frame"))
	}

	_, err := io.ReadFull(relay, fr.Header())
	if err != nil {
		return errors.E(op, err)
	}

	// we have options
	if fr.ReadHL(fr.Header()) > 3 {
		// we should read the options
		optsLen := (fr.ReadHL(fr.Header()) - 3) * frame.WORD
		opts := make([]byte, optsLen)

		// read next part of the frame - options
		_, err = io.ReadFull(relay, opts)
		if err != nil {
			return errors.E(op, err)
		}

		// we should append frame's
		fr.AppendOptions(fr.HeaderPtr(), opts)
	}

	// verify header CRC
	if !fr.VerifyCRC(fr.Header()) {
		return errors.E(op, errors.Str("CRC verification failed"))
	}

	// read the read payload
	pl := fr.ReadPayloadLen(fr.Header())
	// no payload
	if pl == 0 {
		return nil
	}

	pb := make([]byte, pl)
	_, err = io.ReadFull(relay, pb)
	if err != nil {
		return errors.E(op, err)
	}

	fr.WritePayload(pb)
	return nil
}
