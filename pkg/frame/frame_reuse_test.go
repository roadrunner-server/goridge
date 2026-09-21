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

func TestReset_KeepsCapacity(t *testing.T) {
	f := NewFrame()
	sink = f
	f.WriteVersion(f.Header(), Version1)
	f.WriteFlags(f.Header(), CodecJSON)
	f.WriteOptions(f.HeaderPtr(), 7)
	f.WritePayloadLen(f.Header(), 5)
	f.WritePayload([]byte("hello"))
	f.SetStreamFlag(f.Header())
	f.WriteCRC(f.Header())
	payloadCap := cap(f.Payload())

	allocs := testing.AllocsPerRun(1, f.Reset)

	assert.Equal(t, float64(0), allocs)
	assert.Equal(t, 12, len(f.Header()))
	assert.Equal(t, byte(3), f.ReadHL(f.Header()))
	assert.Equal(t, byte(0), f.ReadVersion(f.Header()))
	assert.Equal(t, byte(0), f.ReadFlags())
	assert.Equal(t, uint32(0), f.ReadPayloadLen(f.Header()))
	assert.False(t, f.IsStream(f.Header()))
	assert.Equal(t, 0, len(f.Payload()))
	assert.Equal(t, payloadCap, cap(f.Payload()))
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
