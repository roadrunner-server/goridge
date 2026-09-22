package internal

import (
	"io"
	"net"
	"sync"

	"github.com/roadrunner-server/goridge/v4/internal/bpool"
	"github.com/roadrunner-server/goridge/v4/pkg/frame"
)

// assembleLimit is the frame size, header plus payload, up to which the two are copied into
// one pooled buffer and written with a single call. Above it the copy costs more than the
// second write, so the header and the payload are handed to the writer as a vector instead.
const assembleLimit = 64 << 10

// buffersPool holds two-element net.Buffers so that the vector path does not allocate.
var buffersPool = sync.Pool{
	New: func() any {
		b := make(net.Buffers, 0, 2)
		return &b
	},
}

// SendFrame writes the header and the payload of fr to w.
//
// A frame without a payload is one Write of the header. A frame whose payload is borrowed
// from the pool is one Write of the contiguous slice that Wire returns, at any size. A frame
// over caller memory, as built by From or ReadFrame, is assembled in a pooled buffer up to
// assembleLimit and written once, and above that written as net.Buffers: one writev on a
// net.Conn, the header and then the payload elsewhere. Nothing is copied on that path.
func SendFrame(w io.Writer, fr *frame.Frame) error {
	h, p := fr.Header(), fr.Payload()

	if len(p) == 0 {
		_, err := w.Write(h)
		return err
	}

	// a pooled payload has the header's room in front of it: one write, nothing copied
	if wire, ok := fr.Wire(); ok {
		_, err := w.Write(wire)
		return err
	}

	n := len(h) + len(p)
	if n > assembleLimit {
		return writeVector(w, h, p)
	}

	pb := bpool.Get(n)
	buf := (*pb)[:0]
	buf = append(buf, h...)
	buf = append(buf, p...)
	_, err := w.Write(buf)
	bpool.Put(pb)

	return err
}

// writeVector writes h then p through a pooled net.Buffers. WriteTo consumes the vector by
// re-slicing it, so the full-capacity view is restored and the element references dropped
// before the vector goes back to the pool.
func writeVector(w io.Writer, h, p []byte) error {
	bp := buffersPool.Get().(*net.Buffers)
	v := (*bp)[:0]
	v = append(v, h, p)
	*bp = v

	_, err := bp.WriteTo(w)

	clear(v)
	*bp = v[:0]
	buffersPool.Put(bp)

	return err
}
