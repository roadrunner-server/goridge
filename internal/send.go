package internal

import (
	"io"

	"github.com/roadrunner-server/goridge/v4/pkg/frame"
)

// SendFrame writes the header and the payload of fr to w in one Write call.
// The bytes are assembled in a buffer from the tiered pool, so a frame up to 10 MB does not allocate.
func SendFrame(w io.Writer, fr *frame.Frame) error {
	h, p := fr.Header(), fr.Payload()
	n := uint32(len(h) + len(p)) //nolint:gosec
	pb := get(n)
	buf := (*pb)[:0]
	buf = append(buf, h...)
	buf = append(buf, p...)
	_, err := w.Write(buf)
	put(n, pb)
	return err
}
