package pipe

import (
	"bytes"
	"io"
	"sync"
	"testing"

	"github.com/roadrunner-server/goridge/v4/pkg/frame"
	"github.com/stretchr/testify/assert"
)

const TestPayload = `alsdjf;lskjdgljasg;lkjsalfkjaskldjflkasjdf;lkasjfdalksdjflkajsdf;lfasdgnslsnblna;sldjjfawlkejr;lwjenlksndlfjawl;ejr;lwjelkrjaldfjl;sdjf`

func TestPipeReceive(t *testing.T) {
	pr, pw := io.Pipe()

	relay := NewPipeRelay(pr, pw)

	nf := frame.NewFrame()
	nf.WriteVersion(nf.Header(), frame.Version1)
	nf.WriteFlags(nf.Header(), frame.CONTROL, frame.CodecGob, frame.CodecJSON)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload)))) //nolint:gosec
	nf.WritePayload([]byte(TestPayload))
	nf.WriteCRC(nf.Header())
	assert.Equal(t, true, nf.VerifyCRC(nf.Header()))

	go func(frame *frame.Frame) {
		defer func() {
			_ = pw.Close()
		}()
		err := relay.Send(nf)
		assert.NoError(t, err)
		_ = pw.Close()
	}(nf)

	fr := frame.NewFrame()
	err := relay.Receive(fr)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fr.ReadVersion(fr.Header()), nf.ReadVersion(nf.Header()))
	assert.Equal(t, fr.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, fr.ReadPayloadLen(fr.Header()), nf.ReadPayloadLen(nf.Header()))
	assert.Equal(t, true, fr.VerifyCRC(nf.Header()))
	assert.Equal(t, []byte(TestPayload), fr.Payload())
}

func TestPipeReceiveWithOptions(t *testing.T) {
	pr, pw := io.Pipe()

	relay := NewPipeRelay(pr, pw)

	nf := frame.NewFrame()
	nf.WriteVersion(nf.Header(), frame.Version1)
	nf.WriteFlags(nf.Header(), frame.CONTROL, frame.CodecGob, frame.CodecJSON)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload)))) //nolint:gosec
	nf.WritePayload([]byte(TestPayload))
	nf.WriteOptions(nf.HeaderPtr(), 100, 10000, 100000)
	nf.WriteCRC(nf.Header())
	assert.Equal(t, true, nf.VerifyCRC(nf.Header()))

	go func(frame *frame.Frame) {
		defer func() {
			_ = pw.Close()
		}()
		err := relay.Send(nf)
		assert.NoError(t, err)
		_ = pw.Close()
	}(nf)

	fr := frame.NewFrame()
	err := relay.Receive(fr)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fr.ReadVersion(fr.Header()), nf.ReadVersion(nf.Header()))
	assert.Equal(t, fr.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, fr.ReadPayloadLen(fr.Header()), nf.ReadPayloadLen(nf.Header()))
	assert.Equal(t, true, fr.VerifyCRC(fr.Header()))
	assert.Equal(t, []byte(TestPayload), fr.Payload())
	assert.Equal(t, []uint32{100, 10000, 100000}, fr.ReadOptions(fr.Header()))
}

func TestPipeCRC_Failed(t *testing.T) {
	pr, pw := io.Pipe()

	relay := NewPipeRelay(pr, pw)

	nf := frame.NewFrame()
	nf.WriteVersion(nf.Header(), frame.Version1)
	nf.WriteFlags(nf.Header(), frame.CONTROL)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload)))) //nolint:gosec

	assert.Equal(t, false, nf.VerifyCRC(nf.Header()))

	nf.WritePayload([]byte(TestPayload))

	go func(frame *frame.Frame) {
		defer func() {
			_ = pw.Close()
		}()
		err := relay.Send(nf)
		assert.NoError(t, err)
		_ = pw.Close()
	}(nf)

	fr := frame.NewFrame()
	err := relay.Receive(fr)
	assert.Error(t, err)
	assert.False(t, fr.VerifyCRC(fr.Header()))

	assert.Empty(t, fr.Payload())
}

type discardCloser struct{ io.Writer }

func (discardCloser) Close() error { return nil }

// BenchmarkSendPath mirrors worker.sendFrame in roadrunner-server/pool: a pooled frame and a pooled
// bytes.Buffer, one option, the payload copied into the frame, one Send, then the frame goes back to the pool.
func BenchmarkSendPath(b *testing.B) {
	cases := []struct {
		name string
		size int
	}{
		{name: "1KB", size: 1 << 10},
		{name: "64KB", size: 64 << 10},
		{name: "1MB", size: 1 << 20},
	}
	relay := NewPipeRelay(io.NopCloser(bytes.NewReader(nil)), discardCloser{io.Discard})
	fPool := sync.Pool{New: func() any { return frame.NewFrame() }}
	bPool := sync.Pool{New: func() any { return new(bytes.Buffer) }}
	for _, tc := range cases {
		body := bytes.Repeat([]byte("x"), tc.size)
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(tc.size))
			for b.Loop() {
				fr := fPool.Get().(*frame.Frame)
				buf := bPool.Get().(*bytes.Buffer)
				fr.WriteVersion(fr.Header(), frame.Version1)
				fr.WriteFlags(fr.Header(), frame.CodecRaw)
				buf.Write(body)
				fr.WriteOptions(fr.HeaderPtr(), 0)
				fr.WritePayloadLen(fr.Header(), uint32(buf.Len()))
				fr.WritePayload(buf.Bytes())
				fr.WriteCRC(fr.Header())
				buf.Reset()
				bPool.Put(buf)
				if err := relay.Send(fr); err != nil {
					b.Fatal(err)
				}
				fr.Reset()
				fPool.Put(fr)
			}
		})
	}
}
