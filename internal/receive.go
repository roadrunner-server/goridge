package internal

import (
	"bytes"
	stderr "errors"
	"io"

	"github.com/spiral/errors"
	"github.com/spiral/goridge/v3/pkg/frame"
)

// shortland for the Could not open input file: ../roadrunner/tests/psr-wfsdorker.php
var res = []byte("Could not op") //nolint:gochecknoglobals

func ReceiveFrame(relay io.Reader, fr *frame.Frame) error {
	const op = errors.Op("goridge_frame_receive")

	_, err := io.ReadFull(relay, fr.Header())
	if err != nil {
		return err
	}

	if bytes.Equal(fr.Header(), res) {
		data, errRa := io.ReadAll(relay)
		if errRa == nil && len(data) > 0 {
			return errors.E(op, errors.FileNotFound, errors.Str(string(fr.Header())+string(data)))
		}

		return errors.E(op, errors.FileNotFound, errors.Str("file not found"))
	}

	// we have options
	if fr.ReadHL(fr.Header()) > 3 {
		// we should read the options
		optsLen := (fr.ReadHL(fr.Header()) - 3) * frame.WORD
		opts := make([]byte, optsLen)

		// read next part of the frame - options
		_, err = io.ReadFull(relay, opts)
		if err != nil {
			if stderr.Is(err, io.EOF) {
				return err
			}
			return errors.E(op, err)
		}

		// we should append frame's
		fr.AppendOptions(fr.HeaderPtr(), opts)
	}

	// verify header CRC
	if !fr.VerifyCRC(fr.Header()) {
		return errors.E(op, errors.Errorf("CRC verification failed, bad header: %s", fr.Header()))
	}

	// read the read payload
	pl := fr.ReadPayloadLen(fr.Header())
	// no payload
	if pl == 0 {
		return nil
	}

	pb := get(pl)
	_, err2 := io.ReadFull(relay, (*pb)[:pl])
	if err2 != nil {
		if stderr.Is(err2, io.EOF) {
			put(pl, pb)
			return err
		}
		put(pl, pb)
		return errors.E(op, err2)
	}

	fr.WritePayload((*pb)[:pl])
	put(pl, pb)
	return nil
}
