package internal

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/roadrunner-server/goridge/v4/pkg/frame"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// failReader delivers data[:failAt] normally, then returns err on subsequent reads.
type failReader struct {
	data   []byte
	offset int
	failAt int
	err    error
}

func (r *failReader) Read(p []byte) (int, error) {
	if r.offset >= r.failAt {
		return 0, r.err
	}
	end := min(r.offset+len(p), r.failAt)
	n := copy(p, r.data[r.offset:end])
	r.offset += n
	if r.offset >= r.failAt {
		return n, r.err
	}
	return n, nil
}

// buildValidFrame creates a complete serialized frame with the given payload.
func buildValidFrame(payload []byte) []byte {
	nf := frame.NewFrame()
	nf.WriteVersion(nf.Header(), frame.Version1)
	nf.WriteFlags(nf.Header(), frame.CONTROL)
	nf.WritePayloadLen(nf.Header(), uint32(len(payload))) //nolint:gosec
	nf.WritePayload(payload)
	nf.WriteCRC(nf.Header())
	return nf.Bytes()
}

// buildValidFrameWithOptions creates a frame with options and payload.
func buildValidFrameWithOptions(payload []byte, opts ...uint32) []byte {
	nf := frame.NewFrame()
	nf.WriteVersion(nf.Header(), frame.Version1)
	nf.WriteFlags(nf.Header(), frame.CONTROL)
	nf.WritePayloadLen(nf.Header(), uint32(len(payload))) //nolint:gosec
	nf.WriteOptions(nf.HeaderPtr(), opts...)
	nf.WritePayload(payload)
	nf.WriteCRC(nf.Header())
	return nf.Bytes()
}

func TestReceiveFrame_HeaderReadEOF(t *testing.T) {
	fr := frame.NewFrame()
	err := ReceiveFrame(bytes.NewReader(nil), fr)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF))
}

func TestReceiveFrame_ValidNoPayload(t *testing.T) {
	data := buildValidFrame(nil)
	fr := frame.NewFrame()
	err := ReceiveFrame(bytes.NewReader(data), fr)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(fr.Payload()))
}

func TestReceiveFrame_ValidWithPayload(t *testing.T) {
	payload := []byte("test payload data")
	data := buildValidFrame(payload)
	fr := frame.NewFrame()
	err := ReceiveFrame(bytes.NewReader(data), fr)
	assert.NoError(t, err)
	assert.Equal(t, payload, fr.Payload())
}

func TestReceiveFrame_WithOptions(t *testing.T) {
	payload := []byte("opts payload")
	data := buildValidFrameWithOptions(payload, 42, 12)
	fr := frame.NewFrame()
	err := ReceiveFrame(bytes.NewReader(data), fr)
	assert.NoError(t, err)
	assert.Equal(t, payload, fr.Payload())

	opts := fr.ReadOptions(fr.Header())
	require.Len(t, opts, 2)
	assert.Equal(t, uint32(42), opts[0])
	assert.Equal(t, uint32(12), opts[1])
}

func TestReceiveFrame_CRCFailure(t *testing.T) {
	data := buildValidFrame([]byte("data"))
	// Corrupt a CRC-covered byte (byte 2 is part of payload length, covered by CRC)
	data[2] ^= 0xFF
	fr := frame.NewFrame()
	err := ReceiveFrame(bytes.NewReader(data), fr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestReceiveFrame_OptionsReadEOF(t *testing.T) {
	// Build a valid frame with options, but truncate after the 12-byte header
	// so options bytes can't be read.
	data := buildValidFrameWithOptions([]byte("data"), 1, 2)
	// The 12-byte header has hl>3, so ReceiveFrame will try to read options.
	// Truncate to just the header.
	fr := frame.NewFrame()
	err := ReceiveFrame(bytes.NewReader(data[:12]), fr)
	assert.Error(t, err)
}

func TestReceiveFrame_PayloadEOF(t *testing.T) {
	// Bug 1 regression test:
	// Valid header with payloadLen=100, but EOF when reading payload.
	// receive.go:89 returns `err` (nil from header read) instead of `err2` (EOF).
	//
	// The fix returns io.EOF (zero bytes read from payload → io.ReadFull returns io.EOF, not io.ErrUnexpectedEOF).
	payload := make([]byte, 100)
	data := buildValidFrame(payload)
	// Provide only the 12-byte header (valid) but no payload bytes
	fr := frame.NewFrame()
	err := ReceiveFrame(bytes.NewReader(data[:12]), fr)
	assert.ErrorIs(t, err, io.EOF)
}

func TestReceiveFrame_PayloadNonEOFError(t *testing.T) {
	payload := make([]byte, 50)
	data := buildValidFrame(payload)
	connErr := errors.New("conn reset")
	r := &failReader{
		data:   data,
		failAt: 12, // deliver header, fail on payload
		err:    connErr,
	}
	fr := frame.NewFrame()
	err := ReceiveFrame(r, fr)
	// Non-EOF errors on payload read are wrapped and returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conn reset")
}

func TestReceiveFrame_FileNotFoundMessage(t *testing.T) {
	// When header bytes match "Could not op", ReceiveFrame reads the rest and returns FileNotFound.
	message := []byte("Could not open input file: test.php")
	fr := frame.NewFrame()
	err := ReceiveFrame(bytes.NewReader(message), fr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Could not op")
}

func TestReceiveFrame_FileNotFoundEOF(t *testing.T) {
	// Just "Could not op" (12 bytes) then EOF → generic FileNotFound
	message := []byte("Could not op")
	fr := frame.NewFrame()
	err := ReceiveFrame(bytes.NewReader(message), fr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
}

func FuzzReceiveFrame(f *testing.F) {
	// Seed: valid frame
	f.Add(buildValidFrame([]byte("fuzz seed")))
	// Seed: empty
	f.Add([]byte{})
	// Seed: 12 zero bytes
	f.Add(make([]byte, 12))

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() { recover() }() //nolint:errcheck
		fr := frame.NewFrame()
		_ = ReceiveFrame(bytes.NewReader(data), fr)
	})
}

// BenchmarkReceivePath mirrors worker.receiveFrame in roadrunner-server/pool without the payload clone:
// one frame received from a reused reader, flags and one option read, then the frame reset for reuse.
func BenchmarkReceivePath(b *testing.B) {
	cases := []struct {
		name string
		size int
	}{
		{name: "1KB", size: 1 << 10},
		{name: "64KB", size: 64 << 10},
		{name: "1MB", size: 1 << 20},
	}
	for _, tc := range cases {
		data := buildValidFrameWithOptions(bytes.Repeat([]byte("x"), tc.size), 0)
		b.Run(tc.name, func(b *testing.B) {
			r := bytes.NewReader(data)
			fr := frame.NewFrame()
			b.ReportAllocs()
			b.SetBytes(int64(tc.size))
			for b.Loop() {
				r.Reset(data)
				if err := ReceiveFrame(r, fr); err != nil {
					b.Fatal(err)
				}
				_ = fr.ReadFlags()
				_ = fr.ReadOptions(fr.Header())
				fr.Reset()
			}
		})
	}
}

func TestReceiveFrame_EmptyPayloadClearsPreviousBody(t *testing.T) {
	fr := frame.NewFrame()
	err := ReceiveFrame(bytes.NewReader(buildValidFrame([]byte("previous body"))), fr)
	assert.NoError(t, err)
	assert.Equal(t, []byte("previous body"), fr.Payload())

	// no Reset in between: a payload-less frame must still leave the payload empty
	err = ReceiveFrame(bytes.NewReader(buildValidFrame(nil)), fr)
	assert.NoError(t, err)
	assert.Empty(t, fr.Payload())
}

func TestReceiveFrame_MaxOptionsAndBody(t *testing.T) {
	payload := bytes.Repeat([]byte("q"), 3000)
	data := buildValidFrameWithOptions(payload, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100)
	fr := frame.NewFrame()
	assert.NoError(t, ReceiveFrame(bytes.NewReader(data), fr))
	assert.Equal(t, []uint32{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}, fr.ReadOptions(fr.Header()))
	assert.Equal(t, payload, fr.Payload())
	assert.True(t, fr.VerifyCRC(fr.Header()))
}

func TestReceiveFrame_OptionsReadError(t *testing.T) {
	data := buildValidFrameWithOptions([]byte("body"), 1, 2)
	fr := frame.NewFrame()
	// cut inside the options: 12 header bytes plus 3 of the 8 option bytes
	err := ReceiveFrame(bytes.NewReader(data[:15]), fr)
	// the relay wraps the read error with errors.E, which has no Unwrap, so match the text
	assert.ErrorContains(t, err, io.ErrUnexpectedEOF.Error())
}

// BenchmarkReceivePathWithClone mirrors worker.receiveFrame in roadrunner-server/pool exactly:
// one frame received from a reused reader, flags and the context offset read, body and context
// cloned out because the frame goes back to the pool, then the frame reset.
func BenchmarkReceivePathWithClone(b *testing.B) {
	cases := []struct {
		name string
		size int
	}{
		{name: "200B", size: 200},
		{name: "4KB", size: 4 << 10},
		{name: "64KB", size: 64 << 10},
		{name: "1MB", size: 1 << 20},
	}
	for _, tc := range cases {
		body := bytes.Repeat([]byte("x"), tc.size)
		ctx := []byte(`{"rr":"context"}`)
		data := buildValidFrameWithOptions(append(append([]byte{}, ctx...), body...), uint32(len(ctx))) //nolint:gosec
		b.Run(tc.name, func(b *testing.B) {
			r := bytes.NewReader(data)
			fr := frame.NewFrame()
			var sinkBody, sinkCtx []byte
			b.ReportAllocs()
			b.SetBytes(int64(tc.size))
			for b.Loop() {
				r.Reset(data)
				if err := ReceiveFrame(r, fr); err != nil {
					b.Fatal(err)
				}
				_ = fr.ReadFlags()
				off := fr.ReadOptions(fr.Header())[0]
				sinkBody = bytes.Clone(fr.Payload()[off:])
				sinkCtx = bytes.Clone(fr.Payload()[:off])
				fr.Reset()
			}
			_, _ = sinkBody, sinkCtx
		})
	}
}
