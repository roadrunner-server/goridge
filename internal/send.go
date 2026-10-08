package internal

import (
	"io"
	"net"

	"github.com/roadrunner-server/goridge/v4/pkg/frame"
)

// SendFrame writes the header and the payload of fr to w.
//
// A frame without a payload is one Write of the header. A frame whose payload is borrowed
// from the pool is one Write of the contiguous slice that Wire returns, at any size. A frame
// over caller memory, as built by From or ReadFrame, is written as net.Buffers: one writev on
// a net.Conn, the header and then the payload elsewhere. Nothing is copied on any path.
func SendFrame(w io.Writer, fr *frame.Frame) error {
	h, p := fr.Header(), fr.Payload()

	if len(p) == 0 {
		_, err := w.Write(h)
		return err
	}

	if wire, ok := fr.Wire(); ok {
		_, err := w.Write(wire)
		return err
	}

	bufs := net.Buffers{h, p}
	_, err := bufs.WriteTo(w)
	return err
}
