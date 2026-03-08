package frame

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadHeader_PanicsOnShortInput(t *testing.T) {
	cases := []struct {
		name  string
		input []byte
	}{
		{"nil", nil},
		{"empty", []byte{}},
		{"one_byte", []byte{0x01}},
		{"eleven_bytes", make([]byte, 11)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() { assert.NotNil(t, recover()) }()
			ReadHeader(tc.input)
		})
	}
}

func TestReadHeader_ExactlyTwelveBytes(t *testing.T) {
	data := make([]byte, 12)
	f := ReadHeader(data)
	assert.NotNil(t, f)
	assert.Equal(t, 12, len(f.Header()))
	assert.Nil(t, f.Payload())
}

func TestReadFrame_PanicsOnShortInput(t *testing.T) {
	cases := []struct {
		name  string
		input []byte
	}{
		{"nil", nil},
		{"empty", []byte{}},
		{"one_byte", []byte{0x01}},
		{"eleven_bytes", make([]byte, 11)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() { assert.NotNil(t, recover()) }()
			ReadFrame(tc.input)
		})
	}
}

func TestReadFrame_HighHL_InsufficientData(t *testing.T) {
	// hl=4 means header is 4*WORD=16 bytes. Only 12 bytes provided → panic.
	defer func() { assert.NotNil(t, recover()) }()
	data := make([]byte, 12)
	data[0] = 4 // hl=4
	ReadFrame(data)
}

func TestReadFrame_HighHL_SufficientData(t *testing.T) {
	// hl=4 → header=16 bytes. Provide 20 bytes total → payload=4 bytes.
	data := make([]byte, 20)
	data[0] = 4 // hl=4
	f := ReadFrame(data)
	assert.Equal(t, 16, len(f.Header()))
	assert.Equal(t, 4, len(f.Payload()))
}

func TestReadFrame_ClearsBytes10And11(t *testing.T) {
	// For hl≤3, ReadFrame zeros stream control bytes (10 and 11).
	data := make([]byte, 14)
	data[0] = 3 // hl=3 (default)
	data[10] = 0xFF
	data[11] = 0xFF
	f := ReadFrame(data)
	assert.Equal(t, byte(0), f.Header()[10])
	assert.Equal(t, byte(0), f.Header()[11])
}

func TestWriteVersion_PanicsAbove15(t *testing.T) {
	defer func() { assert.NotNil(t, recover()) }()
	f := NewFrame()
	f.WriteVersion(f.Header(), 16)
}

func TestWriteVersion_MaxValid(t *testing.T) {
	f := NewFrame()
	f.WriteVersion(f.Header(), 15)
	assert.Equal(t, byte(15), f.ReadVersion(f.Header()))
}

func TestWriteVersion_ZeroPreservesHL(t *testing.T) {
	f := NewFrame()
	// default HL is 3
	hlBefore := f.ReadHL(f.Header())
	f.WriteVersion(f.Header(), 0)
	assert.Equal(t, hlBefore, f.ReadHL(f.Header()))
}

func TestReadOptions_PanicsOnShortHeader(t *testing.T) {
	// Craft a header where hl=4 but only 12 bytes exist.
	// ReadOptions will try to read bytes at index 15, causing out-of-bounds panic.
	defer func() { assert.NotNil(t, recover()) }()
	header := make([]byte, 12)
	header[0] = 4 // hl=4 → optionLen=1, needs header[12..15]
	f := NewFrame()
	f.ReadOptions(header)
}

func TestWritePayloadLen_MaxUint32(t *testing.T) {
	f := NewFrame()
	f.WritePayloadLen(f.Header(), math.MaxUint32)
	assert.Equal(t, uint32(math.MaxUint32), f.ReadPayloadLen(f.Header()))
}

func TestWritePayloadLen_Zero(t *testing.T) {
	f := NewFrame()
	f.WritePayloadLen(f.Header(), 0)
	assert.Equal(t, uint32(0), f.ReadPayloadLen(f.Header()))
}

func TestVerifyCRC_PanicsOnShortHeader(t *testing.T) {
	// VerifyCRC needs at least 10 bytes (index 0..9). 6 bytes → panic.
	defer func() { assert.NotNil(t, recover()) }()
	f := NewFrame()
	f.VerifyCRC(make([]byte, 6))
}

func TestCRC_DoesNotCoverStreamFlags(t *testing.T) {
	// CRC is computed over bytes 0..5 only. Mutating byte 10 after WriteCRC should not break verification.
	f := NewFrame()
	f.WriteVersion(f.Header(), Version1)
	f.WriteFlags(f.Header(), CONTROL)
	f.WritePayloadLen(f.Header(), 42)
	f.WriteCRC(f.Header())
	assert.True(t, f.VerifyCRC(f.Header()))

	// Mutate stream control byte — CRC should still pass
	f.Header()[10] = 0xFF
	assert.True(t, f.VerifyCRC(f.Header()))
}

func TestWritePayload_EmptyAndNil(t *testing.T) {
	f := NewFrame()
	f.WritePayload([]byte{})
	assert.Equal(t, 0, len(f.Payload()))

	f2 := NewFrame()
	f2.WritePayload(nil)
	assert.Equal(t, 0, len(f2.Payload()))
}

func TestFrom_NilHeaderPanicsOnAccess(t *testing.T) {
	defer func() { assert.NotNil(t, recover()) }()
	f := From(nil, nil)
	// Any method that indexes into header will panic
	f.ReadVersion(f.Header())
}

func TestWriteFlags_Accumulation(t *testing.T) {
	f := NewFrame()
	f.WriteFlags(f.Header(), CodecJSON)
	f.WriteFlags(f.Header(), CodecGob)
	// CodecJSON=0x08, CodecGob=0x20 → OR'd together = 0x28
	assert.Equal(t, byte(0x28), f.ReadFlags())
}

func TestFrame_ResetClearsEverything(t *testing.T) {
	f := NewFrame()
	f.WriteVersion(f.Header(), Version1)
	f.WriteFlags(f.Header(), CodecJSON)
	f.WritePayloadLen(f.Header(), 999)
	f.WritePayload([]byte("hello"))
	f.WriteCRC(f.Header())

	f.Reset()

	assert.Equal(t, byte(3), f.ReadHL(f.Header()))
	assert.Equal(t, byte(0), f.ReadFlags())
	assert.Equal(t, uint32(0), f.ReadPayloadLen(f.Header()))
	assert.Equal(t, 0, len(f.Payload()))
	assert.Equal(t, byte(0), f.ReadVersion(f.Header()))
}
