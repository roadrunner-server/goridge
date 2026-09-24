//go:build !race

// sync.Pool drops entries at random under the race detector, so pointer identity across Put and Get holds only without it.

package frame

import (
	"bytes"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// singleP pins the test to one P so that a buffer put into a sync.Pool is
// visible to the next Get from the same goroutine.
func singleP(t *testing.T) {
	t.Helper()
	prev := runtime.GOMAXPROCS(1)
	t.Cleanup(func() { runtime.GOMAXPROCS(prev) })
}

// first returns the address of the first byte of b, the identity of its backing array.
func first(b []byte) *byte {
	return &b[:1][0]
}

func TestAllocPayload_ReusesThePooledBuffer(t *testing.T) {
	f := NewFrame()
	sink = f
	f.WritePayload(bytes.Repeat([]byte("a"), 1000))
	addr := first(f.Payload())

	allocs := testing.AllocsPerRun(10, func() { f.WritePayload([]byte("short")) })

	assert.Equal(t, float64(0), allocs)
	assert.Same(t, addr, first(f.Payload()), "a fitting write reuses the same pooled buffer")
}

func TestAllocPayload_GrowthReturnsTheOldBufferToItsTier(t *testing.T) {
	singleP(t)
	f := NewFrame()
	f.WritePayload(bytes.Repeat([]byte("a"), 100)) // 4 KB tier
	old := first(f.Payload())

	f.WritePayload(bytes.Repeat([]byte("b"), tier4K+1)) // 16 KB tier

	got := getBuf(100)
	defer putBuf(got)
	assert.Same(t, old, &(*got)[Headroom], "the 4 KB buffer is back in the 4 KB tier")
	assert.Equal(t, tier16K-Headroom, cap(f.Payload()))
}

func TestReset_ReturnsTheBufferToItsTier(t *testing.T) {
	singleP(t)
	f := NewFrame()
	f.WritePayload(bytes.Repeat([]byte("a"), tier5M-Headroom))
	addr := first(f.Payload())

	f.Reset()

	got := getBuf(tier5M)
	defer putBuf(got)
	assert.Same(t, addr, &(*got)[Headroom], "the 5 MB buffer is back in the 5 MB tier")
	assert.Nil(t, f.Payload())
}

func TestReset_NeverPoolsCallerMemory(t *testing.T) {
	singleP(t)
	// a caller slice with exactly a tier's capacity is the one case Put cannot tell apart by size
	backing := make([]byte, 0, tier4K)
	f := From(make([]byte, 12), backing)
	f.WritePayload([]byte("hello"))
	require.Same(t, first(backing[:1]), first(f.Payload()), "precondition: the payload aliases the caller's memory")

	f.Reset()

	for range 4 {
		got := getBuf(1)
		assert.NotSame(t, first(backing[:1]), first(*got), "the caller's memory must not come out of the pool")
		putBuf(got)
	}
	assert.Equal(t, 0, len(f.Payload()))
}

func TestReset_KeepsTheSmallestTierBuffer(t *testing.T) {
	f := NewFrame()
	sink = f
	f.WritePayload(bytes.Repeat([]byte("a"), 1000))
	addr := first(f.Payload())

	next := bytes.Repeat([]byte("b"), 1000)
	allocs := testing.AllocsPerRun(10, func() {
		f.Reset()
		f.WritePayload(next)
	})

	assert.Equal(t, float64(0), allocs)
	assert.Same(t, addr, first(f.Payload()), "a 4 KB buffer stays on the frame across Reset")
	assert.Equal(t, tier4K-Headroom, cap(f.Payload()))
}

func TestReset_ReturnsTheBufferAFromFrameGrewInto(t *testing.T) {
	singleP(t)
	f := From(make([]byte, 12), make([]byte, 0, 4))
	f.WritePayload(bytes.Repeat([]byte("g"), 5000)) // outgrows the caller's memory into the 16 KB tier
	addr := first(f.Payload())

	f.Reset()

	got := getBuf(5000)
	defer putBuf(got)
	assert.Same(t, addr, &(*got)[Headroom], "the pooled buffer, not the caller's memory, went back to its tier")
}
