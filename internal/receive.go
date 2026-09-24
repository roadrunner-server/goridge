package internal

import (
	"bytes"
	stderr "errors"
	"io"
	"time"

	"github.com/roadrunner-server/errors"
	"github.com/roadrunner-server/goridge/v4/pkg/frame"
)

// shortland for the Could not open input file: ../roadrunner/tests/psr-wfsdorker.php
var res = []byte("Could not op") //nolint:gochecknoglobals

// maxOptionsLen is the longest options block the 4-bit header length can describe.
const maxOptionsLen = (15 - 3) * frame.WORD

// zeroOptions is copied into the header to extend it before the option bytes are read in place. Never written.
var zeroOptions [maxOptionsLen]byte

const validationError = "validation failed on the message sent to STDOUT, see: https://docs.roadrunner.dev/error-codes/stdout-crc, invalid message: %s"

// ReceiveFrame reads a complete frame (header, options, and payload) from the relay into fr.
func ReceiveFrame(relay io.Reader, fr *frame.Frame) error {
	const op = errors.Op("goridge_frame_receive")

	_, err := io.ReadFull(relay, fr.Header())
	if err != nil {
		return err
	}

	// todo: rustatian: think about smarter solution
	if bytes.Equal(fr.Header(), res) {
		data, errRa := io.ReadAll(relay)
		if errRa == nil && len(data) > 0 {
			return errors.E(op, errors.FileNotFound, errors.Str(string(fr.Header())+string(data)))
		}

		return errors.E(op, errors.FileNotFound, errors.Str("file not found"))
	}

	// we have options
	hl := fr.ReadHL(fr.Header())
	if hl > 3 {
		// extend the header by the option bytes and read them in place: the header has
		// spare capacity for the maximum of 10 options, so this does not allocate
		optsLen := int(hl-3) * frame.WORD
		fr.AppendOptions(fr.HeaderPtr(), zeroOptions[:optsLen])
		_, err = io.ReadFull(relay, fr.Header()[12:])
		if err != nil {
			if stderr.Is(err, io.EOF) {
				return err
			}
			return errors.E(op, err)
		}
	}

	// verify header CRC
	if !fr.VerifyCRC(fr.Header()) {
		type deadliner interface {
			SetReadDeadline(time.Time) error
		}

		if d, ok := relay.(deadliner); ok {
			err = d.SetReadDeadline(time.Now().Add(time.Second * 2))
			if err != nil {
				return errors.E(op, errors.Errorf(validationError, fr.Header()))
			}

			// we don't care about error here
			resp, _ := io.ReadAll(relay)

			return errors.E(op, errors.Errorf(validationError, string(fr.Header())+string(resp)))
		}

		// no deadline, so, only 14 bytes
		return errors.E(op, errors.Errorf(validationError, fr.Header()))
	}

	// read the payload straight into the frame's buffer; on error the buffer stays
	// on the frame and the caller's Reset returns it to the pool
	pl := fr.ReadPayloadLen(fr.Header())
	_, err = io.ReadFull(relay, fr.AllocPayload(int(pl)))
	if err != nil {
		if stderr.Is(err, io.EOF) {
			return err
		}
		return errors.E(op, err)
	}

	return nil
}
