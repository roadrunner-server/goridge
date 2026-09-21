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

func TestSendFrame_DoesNotAllocate(t *testing.T) {
	if raceEnabled {
		t.Skip("sync.Pool drops entries at random under the race detector")
	}
	fr := buildTestFrame(bytes.Repeat([]byte("x"), 1024), 1)
	allocs := testing.AllocsPerRun(100, func() {
		if err := SendFrame(io.Discard, fr); err != nil {
			t.Fatal(err)
		}
	})
	assert.Equal(t, float64(0), allocs)
}
