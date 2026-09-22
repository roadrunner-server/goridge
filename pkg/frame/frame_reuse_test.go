package frame

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// sink keeps frames on the heap so that AllocsPerRun counts their buffers.
var sink *Frame

func TestWritePayload_ReusesCapacity(t *testing.T) {
	cases := []struct {
		name       string
		first      []byte
		second     []byte
		wantAllocs float64
	}{
		{name: "shorter_data_fits", first: []byte("hello world"), second: []byte("hi"), wantAllocs: 0},
		{name: "same_length_data_fits", first: []byte("hello"), second: []byte("world"), wantAllocs: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := NewFrame()
			sink = f
			f.WritePayload(tc.first)
			allocs := testing.AllocsPerRun(1, func() { f.WritePayload(tc.second) })
			assert.Equal(t, tc.wantAllocs, allocs)
			assert.Equal(t, tc.second, f.Payload())
		})
	}
}

func TestWritePayload_GrowsBeyondCapacity(t *testing.T) {
	f := NewFrame()
	f.WritePayload([]byte("hi"))
	data := bytes.Repeat([]byte("abcdefghij"), 20)
	f.WritePayload(data)
	assert.Equal(t, data, f.Payload())
}

func TestReset_ReleasesPayload(t *testing.T) {
	f := NewFrame()
	sink = f
	f.WriteVersion(f.Header(), Version1)
	f.WriteFlags(f.Header(), CodecJSON)
	f.WriteOptions(f.HeaderPtr(), 7)
	f.WritePayloadLen(f.Header(), 512)
	f.WritePayload(bytes.Repeat([]byte("x"), 512))
	f.SetStreamFlag(f.Header())
	f.WriteCRC(f.Header())

	allocs := testing.AllocsPerRun(1, f.Reset)

	assert.Equal(t, float64(0), allocs)
	assert.Equal(t, 12, len(f.Header()))
	assert.Equal(t, byte(3), f.ReadHL(f.Header()))
	assert.Equal(t, byte(0), f.ReadVersion(f.Header()))
	assert.Equal(t, byte(0), f.ReadFlags())
	assert.Equal(t, uint32(0), f.ReadPayloadLen(f.Header()))
	assert.False(t, f.IsStream(f.Header()))
	assert.Empty(t, f.Payload(), "a reset frame has an empty payload")
	assert.NotPanics(t, f.Reset, "a second Reset has nothing to release")
}

func TestReset_ReleasesBuffersAboveTheSmallestTier(t *testing.T) {
	f := NewFrame()
	f.WritePayload(bytes.Repeat([]byte("x"), 5000)) // 16 KB tier
	f.Reset()
	assert.Nil(t, f.Payload(), "a buffer above the smallest tier goes back to the pool")
}

func TestNewFrame_HasNoPayloadBuffer(t *testing.T) {
	f := NewFrame()
	assert.Nil(t, f.Payload())
	assert.Equal(t, 0, cap(f.Payload()))
}

func TestAllocPayload_ZeroNeverTakesABuffer(t *testing.T) {
	f := NewFrame()
	sink = f
	allocs := testing.AllocsPerRun(1, func() { f.AllocPayload(0) })
	assert.Equal(t, float64(0), allocs)
	assert.Nil(t, f.Payload())
}

func TestAllocPayload_ReturnsWritableSliceOfRequestedLength(t *testing.T) {
	f := NewFrame()
	buf := f.AllocPayload(300)
	assert.Equal(t, 300, len(buf))
	copy(buf, bytes.Repeat([]byte("y"), 300))
	assert.Equal(t, bytes.Repeat([]byte("y"), 300), f.Payload())
}

func TestWritePayload_FromFrameWritesInPlaceWhenItFits(t *testing.T) {
	backing := make([]byte, 0, 64)
	f := From(make([]byte, 12), backing)
	f.WritePayload([]byte("hello"))
	assert.Equal(t, []byte("hello"), backing[:5], "a fitting write goes into the caller's memory, as before")
	assert.Equal(t, []byte("hello"), f.Payload())
}

func TestWritePayload_FromFrameGrowsIntoAPooledBuffer(t *testing.T) {
	backing := make([]byte, 0, 4)
	f := From(make([]byte, 12), backing)
	data := bytes.Repeat([]byte("z"), 100)
	f.WritePayload(data)
	assert.Equal(t, data, f.Payload())
	assert.Equal(t, 0, len(backing), "the caller's memory is untouched once the payload outgrows it")
	f.Reset()
	assert.Empty(t, f.Payload())
}

func TestWriteOptions_ReusesHeaderCapacity(t *testing.T) {
	cases := []struct {
		name       string
		options    []uint32
		wantAllocs float64
	}{
		{name: "one_option", options: []uint32{1}, wantAllocs: 0},
		{name: "two_options", options: []uint32{1, 2}, wantAllocs: 0},
		{name: "ten_options", options: []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, wantAllocs: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := NewFrame()
			sink = f
			allocs := testing.AllocsPerRun(1, func() {
				f.Reset()
				f.WriteOptions(f.HeaderPtr(), tc.options...)
			})
			assert.Equal(t, tc.wantAllocs, allocs)
			assert.Equal(t, tc.options, f.ReadOptions(f.Header()))
		})
	}
}

func TestWriteOptions_AfterResetClearsStaleOptions(t *testing.T) {
	f := NewFrame()
	f.WriteOptions(f.HeaderPtr(), 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF)
	f.Reset()
	f.WriteOptions(f.HeaderPtr(), 9)
	assert.Equal(t, 16, len(f.Header()))
	assert.Equal(t, []uint32{9}, f.ReadOptions(f.Header()))
}

func TestWriteOptions_FromFrameAllocatesWhenCapacityIsShort(t *testing.T) {
	header := make([]byte, 12)
	header[0] = 3
	want := bytes.Clone(header)
	f := From(header, nil)
	f.WriteOptions(f.HeaderPtr(), 5)
	assert.Equal(t, 16, len(f.Header()))
	assert.Equal(t, []uint32{5}, f.ReadOptions(f.Header()))
	assert.Equal(t, want, header)
}

func TestAppendOptions_ReusesHeaderCapacity(t *testing.T) {
	f := NewFrame()
	sink = f
	opts := []byte{1, 0, 0, 0, 2, 0, 0, 0}
	allocs := testing.AllocsPerRun(1, func() {
		f.Reset()
		f.AppendOptions(f.HeaderPtr(), opts)
	})
	assert.Equal(t, float64(0), allocs)
	assert.Equal(t, 20, len(f.Header()))
	assert.Equal(t, opts, f.Header()[12:])
}

func TestWire_IsHeaderThenPayload(t *testing.T) {
	cases := []struct {
		name    string
		opts    []uint32
		payload []byte
	}{
		{name: "no_options", payload: []byte("payload")},
		{name: "one_option", opts: []uint32{42}, payload: bytes.Repeat([]byte("p"), 3000)},
		{name: "ten_options", opts: []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, payload: bytes.Repeat([]byte("q"), 70000)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := NewFrame()
			f.WriteVersion(f.Header(), Version1)
			f.WriteOptions(f.HeaderPtr(), tc.opts...)
			f.WritePayloadLen(f.Header(), uint32(len(tc.payload))) //nolint:gosec
			f.WritePayload(tc.payload)
			f.WriteCRC(f.Header())

			wire, ok := f.Wire()

			assert.True(t, ok)
			assert.Equal(t, f.Bytes(), wire)
			assert.Equal(t, tc.payload, f.Payload(), "the payload is untouched")
		})
	}
}

func TestWire_FalseWithoutAPooledPayload(t *testing.T) {
	empty := NewFrame()
	_, ok := empty.Wire()
	assert.False(t, ok, "nothing to send as one slice without a payload")

	aliased := From(make([]byte, 12), []byte("caller memory"))
	_, ok = aliased.Wire()
	assert.False(t, ok, "caller memory has no headroom in front of it")
}

func TestWire_DoesNotAllocate(t *testing.T) {
	f := NewFrame()
	sink = f
	f.WriteOptions(f.HeaderPtr(), 7)
	f.WritePayload(bytes.Repeat([]byte("w"), 1000))
	allocs := testing.AllocsPerRun(10, func() { _, _ = f.Wire() })
	assert.Equal(t, float64(0), allocs)
}

func TestAllocPayload_LeavesHeadroomInThePooledBuffer(t *testing.T) {
	f := NewFrame()
	f.WritePayload([]byte("x"))
	assert.Equal(t, Headroom, cap(*f.pb)-cap(f.Payload()), "the payload starts Headroom bytes into the buffer")
}
