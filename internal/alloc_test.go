//go:build !race

// sync.Pool drops entries at random under the race detector, so the allocation counts hold only without it.

package internal

import (
	"bytes"
	"io"
	"runtime"
	"testing"

	"github.com/roadrunner-server/goridge/v4/internal/bpool"
	"github.com/roadrunner-server/goridge/v4/pkg/frame"
	"github.com/stretchr/testify/assert"
)

func TestSendFrame_DoesNotAllocate(t *testing.T) {
	fr := buildTestFrame(bytes.Repeat([]byte("x"), 1024), 1)
	allocs := testing.AllocsPerRun(100, func() {
		if err := SendFrame(io.Discard, fr); err != nil {
			t.Fatal(err)
		}
	})
	assert.Equal(t, float64(0), allocs)
}

func TestReceiveFrame_DoesNotAllocateOnReuse(t *testing.T) {
	payload := bytes.Repeat([]byte("x"), 1024)
	data := buildValidFrameWithOptions(payload, 42, 12)
	r := bytes.NewReader(data)
	fr := frame.NewFrame()
	allocs := testing.AllocsPerRun(100, func() {
		fr.Reset()
		r.Reset(data)
		if err := ReceiveFrame(r, fr); err != nil {
			t.Fatal(err)
		}
	})
	assert.Equal(t, float64(0), allocs)
	assert.Equal(t, []uint32{42, 12}, fr.ReadOptions(fr.Header()))
	assert.Equal(t, payload, fr.Payload())
}

// singleP pins the test to one P so that a buffer put into a sync.Pool is
// visible to the next Get from the same goroutine.
func singleP(t *testing.T) {
	t.Helper()
	prev := runtime.GOMAXPROCS(1)
	t.Cleanup(func() { runtime.GOMAXPROCS(prev) })
}

func TestReceiveFrame_OptionsDoNotTouchThePool(t *testing.T) {
	// ten options, the maximum, and no payload: nothing here may allocate or take a buffer
	data := buildValidFrameWithOptions(nil, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	r := bytes.NewReader(data)
	fr := frame.NewFrame()
	allocs := testing.AllocsPerRun(100, func() {
		fr.Reset()
		r.Reset(data)
		if err := ReceiveFrame(r, fr); err != nil {
			t.Fatal(err)
		}
	})
	assert.Equal(t, float64(0), allocs)
	assert.Equal(t, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, fr.ReadOptions(fr.Header()))
	assert.Empty(t, fr.Payload())
}

func TestReceiveFrame_PayloadLivesInThePooledBuffer(t *testing.T) {
	singleP(t)
	payload := bytes.Repeat([]byte("p"), int(bpool.FiveMB)-frame.Headroom)
	data := buildValidFrameWithOptions(payload, 0)
	fr := frame.NewFrame()
	if err := ReceiveFrame(bytes.NewReader(data), fr); err != nil {
		t.Fatal(err)
	}
	addr := &fr.Payload()[0]
	assert.Equal(t, int(bpool.FiveMB)-frame.Headroom, cap(fr.Payload()), "the body was read into a 5 MB tier buffer, behind the headroom")

	fr.Reset()

	got := bpool.Get(bpool.FiveMB)
	defer bpool.Put(got)
	assert.Same(t, addr, &(*got)[frame.Headroom], "after Reset the buffer is back in its tier, not stuck on the frame")
	assert.Nil(t, fr.Payload())
}

func TestReceiveFrame_PartialBodyIsReleasedByReset(t *testing.T) {
	singleP(t)
	payload := bytes.Repeat([]byte("p"), int(bpool.SixtyFourKB)-frame.Headroom)
	data := buildValidFrameWithOptions(payload, 0)
	fr := frame.NewFrame()

	err := ReceiveFrame(bytes.NewReader(data[:len(data)-100]), fr)
	assert.ErrorContains(t, err, io.ErrUnexpectedEOF.Error())
	addr := &fr.AllocPayload(1)[0] // the buffer taken for the body is still on the frame

	fr.Reset()

	got := bpool.Get(bpool.SixtyFourKB)
	defer bpool.Put(got)
	assert.Same(t, addr, &(*got)[frame.Headroom], "the buffer of the failed read went back to its tier")
}

func TestSendFrame_LargeFrameDoesNotAllocate(t *testing.T) {
	fr := buildTestFrame(bytes.Repeat([]byte("x"), 1<<20), 1)
	allocs := testing.AllocsPerRun(100, func() {
		if err := SendFrame(io.Discard, fr); err != nil {
			t.Fatal(err)
		}
	})
	assert.Equal(t, float64(0), allocs)
}

func TestSendFrame_AliasedLargeFrameDoesNotAllocate(t *testing.T) {
	fr := aliasedCopy(buildTestFrame(bytes.Repeat([]byte("x"), 1<<20), 1))
	allocs := testing.AllocsPerRun(100, func() {
		if err := SendFrame(io.Discard, fr); err != nil {
			t.Fatal(err)
		}
	})
	assert.Equal(t, float64(0), allocs)
}
