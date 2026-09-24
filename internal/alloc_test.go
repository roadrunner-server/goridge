//go:build !race

// sync.Pool drops entries at random under the race detector, so the allocation counts hold only without it.

package internal

import (
	"bytes"
	"io"
	"testing"

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
