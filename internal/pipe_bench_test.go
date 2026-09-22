package internal

import (
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"testing"

	"github.com/roadrunner-server/goridge/v4/internal/bpool"
	"github.com/roadrunner-server/goridge/v4/pkg/frame"
)

// sendAssembled is the pre-vector strategy for any size: copy header and payload into one
// pooled buffer and write once. Kept here only to measure it against writeVector.
func sendAssembled(w io.Writer, fr *frame.Frame) error {
	h, p := fr.Header(), fr.Payload()
	n := len(h) + len(p)
	pb := bpool.Get(n)
	buf := (*pb)[:0]
	buf = append(buf, h...)
	buf = append(buf, p...)
	_, err := w.Write(buf)
	bpool.Put(pb)
	return err
}

func sendVector(w io.Writer, fr *frame.Frame) error {
	return writeVector(w, fr.Header(), fr.Payload())
}

var strategySizes = []struct {
	name string
	size int
}{
	{"16KB", 16 << 10},
	{sixtyFourKB, 64 << 10},
	{"256KB", 256 << 10},
	{oneMB, 1 << 20},
}

var strategies = []struct {
	name string
	send func(io.Writer, *frame.Frame) error
}{
	{"assemble", sendAssembled},
	{"vector", sendVector},
}

func benchStrategies(b *testing.B, w io.Writer) {
	b.Helper()
	for _, sz := range strategySizes {
		fr := buildTestFrame(bytes.Repeat([]byte("x"), sz.size), 1)
		for _, st := range strategies {
			b.Run(st.name+"/"+sz.name, func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(sz.size))
				for b.Loop() {
					if err := st.send(w, fr); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

// BenchmarkPipeSendStrategy sends frames through a real pipe with a draining reader on the
// other end and compares one memcpy plus one write against two writes and no copy.
func BenchmarkPipeSendStrategy(b *testing.B) {
	pr, pw, err := os.Pipe()
	if err != nil {
		b.Fatal(err)
	}
	defer pr.Close()
	defer pw.Close()
	go func() { _, _ = io.Copy(io.Discard, pr) }()

	benchStrategies(b, pw)
}

// BenchmarkSocketSendStrategy is the same comparison over a TCP loopback, where vector means writev.
func BenchmarkSocketSendStrategy(b *testing.B) {
	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer ln.Close()
	go func() {
		conn, aErr := ln.Accept()
		if aErr != nil {
			return
		}
		defer conn.Close()
		_, _ = io.Copy(io.Discard, conn)
	}()
	conn, err := (&net.Dialer{}).DialContext(context.Background(), "tcp", ln.Addr().String())
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	benchStrategies(b, conn)
}
