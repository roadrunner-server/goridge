package internal

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBufferPool_GetReturnsSmallestFittingTier(t *testing.T) {
	Preallocate()
	cases := []struct {
		name string
		size uint32
		tier uint32
	}{
		{"one_byte", 1, FourKB},
		{"200B", 200, FourKB},
		{"exactly_4KB", FourKB, FourKB},
		{"over_4KB", FourKB + 1, SixteenKB},
		{"exactly_16KB", SixteenKB, SixteenKB},
		{"over_16KB", SixteenKB + 1, SixtyFourKB},
		{"exactly_64KB", SixtyFourKB, SixtyFourKB},
		{"over_64KB", SixtyFourKB + 1, TwoFiftySixKB},
		{"exactly_256KB", TwoFiftySixKB, TwoFiftySixKB},
		{"over_256KB", TwoFiftySixKB + 1, OneMB},
		{"exactly_1MB", OneMB, OneMB},
		{"over_1MB", OneMB + 1, FiveMB},
		{"exactly_5MB", FiveMB, FiveMB},
		{"over_5MB", FiveMB + 1, TenMB},
		{"exactly_10MB", TenMB, TenMB},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := get(tc.size)
			assert.Equal(t, int(tc.tier), cap(*buf), "capacity must be the tier size")
			assert.GreaterOrEqual(t, len(*buf), int(tc.size), "length must cover the request")
			put(buf)
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
func drainTier(size uint32, n int) []*[]byte {
	bufs := make([]*[]byte, 0, n)
	for range n {
		bufs = append(bufs, get(size))
	}
	return bufs
}

func TestBufferPool_OversizedBufferIsNotPooled(t *testing.T) {
	Preallocate()
	singleP(t)

	size := TenMB + 1
	buf := get(size)
	assert.Equal(t, int(size), len(*buf))
	put(buf)

	bufs := drainTier(TenMB, 4)
	for _, b := range bufs {
		assert.Equal(t, int(TenMB), cap(*b), "the 10 MB tier must only hand out 10 MB buffers")
	}
	for _, b := range bufs {
		put(b)
	}
}

func TestBufferPool_ForeignBufferIsNotPooled(t *testing.T) {
	Preallocate()
	singleP(t)

	foreign := make([]byte, 100)
	put(&foreign)

	bufs := drainTier(1, 4)
	for _, b := range bufs {
		assert.Equal(t, int(FourKB), cap(*b), "the 4 KB tier must only hand out 4 KB buffers")
	}
	for _, b := range bufs {
		put(b)
	}
}

func TestBufferPool_PutReturnsBufferToItsOwnTier(t *testing.T) {
	Preallocate()
	singleP(t)

	// routing put by the requested size would send this 4 KB buffer to the 1 MB tier,
	// and the next 1 MB request would get a buffer it does not fit into
	small := get(200)
	put(small)

	for _, b := range drainTier(OneMB, 4) {
		assert.Equal(t, int(OneMB), cap(*b), "the 1 MB tier must only hand out 1 MB buffers")
		put(b)
	}
}
