package bpool

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet_ReturnsSmallestFittingTier(t *testing.T) {
	cases := []struct {
		name string
		size int
		tier uint32
	}{
		{"one_byte", 1, FourKB},
		{"200B", 200, FourKB},
		{"exactly_4KB", int(FourKB), FourKB},
		{"over_4KB", int(FourKB) + 1, SixteenKB},
		{"exactly_16KB", int(SixteenKB), SixteenKB},
		{"over_16KB", int(SixteenKB) + 1, SixtyFourKB},
		{"exactly_64KB", int(SixtyFourKB), SixtyFourKB},
		{"over_64KB", int(SixtyFourKB) + 1, TwoFiftySixKB},
		{"exactly_256KB", int(TwoFiftySixKB), TwoFiftySixKB},
		{"over_256KB", int(TwoFiftySixKB) + 1, OneMB},
		{"exactly_1MB", int(OneMB), OneMB},
		{"over_1MB", int(OneMB) + 1, FiveMB},
		{"exactly_5MB", int(FiveMB), FiveMB},
		{"over_5MB", int(FiveMB) + 1, TenMB},
		{"exactly_10MB", int(TenMB), TenMB},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := Get(tc.size)
			assert.Equal(t, int(tc.tier), cap(*buf), "capacity must be the tier size")
			assert.GreaterOrEqual(t, len(*buf), tc.size, "length must cover the request")
			Put(buf)
		})
	}
}

// singleP pins the test to one P so that a buffer put into a sync.Pool is
// visible to the next get from the same goroutine.
func singleP(t *testing.T) {
	t.Helper()
	prev := runtime.GOMAXPROCS(1)
	t.Cleanup(func() { runtime.GOMAXPROCS(prev) })
}

// drainTier takes n buffers from the tier serving size. Under singleP the most
// recently put buffer, if it was pooled at all, is among the first two.
func drainTier(size int, n int) []*[]byte {
	bufs := make([]*[]byte, 0, n)
	for range n {
		bufs = append(bufs, Get(size))
	}
	return bufs
}

func TestPut_OversizedBufferIsNotPooled(t *testing.T) {
	singleP(t)

	size := int(TenMB) + 1
	buf := Get(size)
	assert.Equal(t, size, len(*buf))
	Put(buf)

	bufs := drainTier(int(TenMB), 4)
	for _, b := range bufs {
		assert.Equal(t, int(TenMB), cap(*b), "the 10 MB tier must only hand out 10 MB buffers")
	}
	for _, b := range bufs {
		Put(b)
	}
}

func TestPut_ForeignBufferIsNotPooled(t *testing.T) {
	singleP(t)

	foreign := make([]byte, 100)
	Put(&foreign)

	bufs := drainTier(1, 4)
	for _, b := range bufs {
		assert.Equal(t, int(FourKB), cap(*b), "the 4 KB tier must only hand out 4 KB buffers")
	}
	for _, b := range bufs {
		Put(b)
	}
}

func TestPut_ReturnsBufferToItsOwnTier(t *testing.T) {
	singleP(t)

	// routing put by the requested size would send this 4 KB buffer to the 1 MB tier,
	// and the next 1 MB request would get a buffer it does not fit into
	small := Get(200)
	Put(small)

	for _, b := range drainTier(int(OneMB), 4) {
		assert.Equal(t, int(OneMB), cap(*b), "the 1 MB tier must only hand out 1 MB buffers")
		Put(b)
	}
}
