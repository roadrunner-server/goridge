package internal

import (
	"bytes"
	"errors"
	"testing"

	"github.com/roadrunner-server/goridge/v4/pkg/frame"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failWriter struct{ err error }

func (w failWriter) Write([]byte) (int, error) { return 0, w.err }

// buildTestFrame returns a frame with the given payload and options, ready to send.
func buildTestFrame(payload []byte, opts ...uint32) *frame.Frame {
	nf := frame.NewFrame()
	nf.WriteVersion(nf.Header(), frame.Version1)
	nf.WriteFlags(nf.Header(), frame.CodecRaw)
	nf.WriteOptions(nf.HeaderPtr(), opts...)
	nf.WritePayloadLen(nf.Header(), uint32(len(payload))) //nolint:gosec
	nf.WritePayload(payload)
	nf.WriteCRC(nf.Header())
	return nf
}

func TestSendFrame_WritesFrameBytes(t *testing.T) {
	cases := []struct {
		name    string
		opts    []uint32
		payload []byte
	}{
		{name: "header_only"},
		{name: "header_and_options", opts: []uint32{42, 12}},
		{name: "header_and_payload", payload: []byte("payload")},
		{name: "header_options_and_payload", opts: []uint32{42, 12}, payload: []byte("payload")},
		{name: "exactly_at_limit", payload: bytes.Repeat([]byte("l"), assembleLimit-12)},
		{name: "one_over_limit", payload: bytes.Repeat([]byte("m"), assembleLimit-12+1)},
		{name: "one_over_limit_with_options", opts: []uint32{7}, payload: bytes.Repeat([]byte("n"), assembleLimit-16+1)},
		{name: oneMB, payload: bytes.Repeat([]byte("o"), 1<<20)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fr := buildTestFrame(tc.payload, tc.opts...)
			var out bytes.Buffer
			require.NoError(t, SendFrame(&out, fr))
			assert.Equal(t, fr.Bytes(), out.Bytes())

			rf := frame.NewFrame()
			require.NoError(t, ReceiveFrame(bytes.NewReader(out.Bytes()), rf))
			assert.Equal(t, tc.opts, rf.ReadOptions(rf.Header()))
			assert.Equal(t, len(tc.payload), len(rf.Payload()))
			assert.Equal(t, string(tc.payload), string(rf.Payload()))
		})
	}
}

func TestSendFrame_ReturnsWriterError(t *testing.T) {
	wantErr := errors.New("write failed")
	err := SendFrame(failWriter{err: wantErr}, buildTestFrame([]byte("payload")))
	assert.ErrorIs(t, err, wantErr)
}

// countingWriter records every Write it receives.
type countingWriter struct {
	writes int
	bytes  int
}

func (w *countingWriter) Write(p []byte) (int, error) {
	w.writes++
	w.bytes += len(p)
	return len(p), nil
}

func TestSendFrame_SmallFrameIsOneWrite(t *testing.T) {
	w := &countingWriter{}
	fr := buildTestFrame(bytes.Repeat([]byte("s"), 1024), 1)
	require.NoError(t, SendFrame(w, fr))
	assert.Equal(t, 1, w.writes)
	assert.Equal(t, len(fr.Bytes()), w.bytes)
}

func TestSendFrame_HeaderOnlyFrameIsOneWrite(t *testing.T) {
	w := &countingWriter{}
	fr := buildTestFrame(nil, 1)
	require.NoError(t, SendFrame(w, fr))
	assert.Equal(t, 1, w.writes)
	assert.Equal(t, len(fr.Header()), w.bytes)
}

func TestSendFrame_LargeFrameOnPlainWriterIsHeaderThenPayload(t *testing.T) {
	// a writer without writev support gets the header and the payload as two writes, nothing copied
	w := &countingWriter{}
	fr := buildTestFrame(bytes.Repeat([]byte("s"), 1<<20), 1)
	require.NoError(t, SendFrame(w, fr))
	assert.Equal(t, 2, w.writes)
	assert.Equal(t, len(fr.Bytes()), w.bytes)
}

// failAfterWriter fails on the write with the given ordinal.
type failAfterWriter struct {
	failOn int
	n      int
	err    error
}

func (w *failAfterWriter) Write(p []byte) (int, error) {
	w.n++
	if w.n == w.failOn {
		return 0, w.err
	}
	return len(p), nil
}

func TestSendFrame_ReturnsWriterErrorOnTheLargePath(t *testing.T) {
	wantErr := errors.New("write failed")
	fr := buildTestFrame(bytes.Repeat([]byte("s"), 1<<20), 1)
	err := SendFrame(&failAfterWriter{failOn: 2, err: wantErr}, fr)
	assert.ErrorIs(t, err, wantErr)
}
