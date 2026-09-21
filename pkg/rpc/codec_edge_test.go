package rpc

import (
	"bytes"
	"io"
	"net/rpc"
	"testing"

	"github.com/roadrunner-server/goridge/v4/pkg/frame"
	"github.com/stretchr/testify/assert"
)

// nopCloserRWC wraps an io.ReadWriter with a no-op Close method.
type nopCloserRWC struct {
	io.ReadWriter
}

func (nopCloserRWC) Close() error { return nil }

func TestWriteResponse_NilCodecPanic_Bug2(t *testing.T) {
	// Bug 2 regression test:
	// When Seq is not in the sync.Map and r.Error is empty,
	// WriteResponse reaches the switch at codec.go:97 where codec is nil.
	// The type assertion codec.(byte) panics on nil interface.
	defer func() {
		r := recover()
		assert.Nil(t, r, "WriteResponse panicked on nil codec, expected no panic")
	}()

	buf := &bytes.Buffer{}
	c := NewCodec(nopCloserRWC{buf})

	resp := &rpc.Response{
		ServiceMethod: "Test.Method",
		Seq:           99999, // Seq not stored in sync.Map
		Error:         "",    // No error → doesn't short-circuit to handleError
	}

	// Previously panicked at codec.(byte) due to LoadAndDelete returning nil; ensure no panic now.
	_ = c.WriteResponse(resp, "some body")
}

func TestWriteResponse_ErrorPath_NilCodecSafe(t *testing.T) {
	// When Seq is not in sync.Map but r.Error is non-empty,
	// the code takes the handleError path before reaching the switch.
	// This should not panic.
	buf := &bytes.Buffer{}
	c := NewCodec(nopCloserRWC{buf})

	resp := &rpc.Response{
		ServiceMethod: "Test.Method",
		Seq:           88888,
		Error:         "some error occurred",
	}

	// Should not panic — error path short-circuits before nil codec assertion
	err := c.WriteResponse(resp, nil)
	// handleError returns an error wrapping r.Error
	assert.Error(t, err)
}

func TestCodec_DoubleClose(t *testing.T) {
	buf := &bytes.Buffer{}
	c := NewCodec(nopCloserRWC{buf})

	err := c.Close()
	assert.NoError(t, err)

	// Second close should return nil
	err = c.Close()
	assert.NoError(t, err)
}

func TestClientCodec_DoubleClose(t *testing.T) {
	buf := &bytes.Buffer{}
	c := NewClientCodec(nopCloserRWC{buf})

	err := c.Close()
	assert.NoError(t, err)

	// Second close should return nil
	err = c.Close()
	assert.NoError(t, err)
}

func TestReadRequestBody_NilOut(t *testing.T) {
	buf := &bytes.Buffer{}
	c := NewCodec(nopCloserRWC{buf})

	// ReadRequestBody with nil out should return nil immediately
	err := c.ReadRequestBody(nil)
	assert.NoError(t, err)
}

func TestReadResponseBody_NilOut(t *testing.T) {
	buf := &bytes.Buffer{}
	c := NewClientCodec(nopCloserRWC{buf})

	// Set a valid frame so putFrame in defer doesn't panic
	c.frame = frame.NewFrame()

	err := c.ReadResponseBody(nil)
	assert.NoError(t, err)
}

func TestStoreCodec_AllCodecs(t *testing.T) {
	cases := []struct {
		name    string
		flag    byte
		wantErr bool
	}{
		{"proto", frame.CodecProto, false},
		{"json", frame.CodecJSON, false},
		{"raw", frame.CodecRaw, false},
		{"gob", frame.CodecGob, false},
		{"msgpack", frame.CodecMsgpack, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			c := NewCodec(nopCloserRWC{buf})

			req := &rpc.Request{Seq: 1}
			err := c.storeCodec(req, tc.flag)

			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			val, ok := c.codec.Load(req.Seq)
			assert.True(t, ok)
			assert.Equal(t, tc.flag, val.(byte))
		})
	}
}

// responseFrame returns the wire bytes of a response frame with the given payload and options.
func responseFrame(payload []byte, opts ...uint32) []byte {
	fr := frame.NewFrame()
	fr.WriteVersion(fr.Header(), frame.Version1)
	fr.WriteFlags(fr.Header(), frame.CodecRaw)
	fr.WriteOptions(fr.HeaderPtr(), opts...)
	fr.WritePayloadLen(fr.Header(), uint32(len(payload))) //nolint:gosec
	fr.WritePayload(payload)
	fr.WriteCRC(fr.Header())
	return fr.Bytes()
}

func TestReadResponseHeader_ErrorLeavesNoFrame(t *testing.T) {
	cases := []struct {
		name    string
		data    []byte
		wantErr string
	}{
		{name: "truncated_frame", data: responseFrame([]byte("abc"), 1, 3)[:12], wantErr: "EOF"},
		{name: "one_option", data: responseFrame([]byte("abc"), 1), wantErr: errOpts},
		{name: "offset_exceeds_payload", data: responseFrame([]byte("abc"), 1, 100), wantErr: "method length offset exceeds payload bounds"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewClientCodec(nopCloserRWC{bytes.NewBuffer(tc.data)})
			err := c.ReadResponseHeader(&rpc.Response{})
			assert.ErrorContains(t, err, tc.wantErr)
			assert.Nil(t, c.frame)
		})
	}
}
