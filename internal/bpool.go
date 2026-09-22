package internal

import (
	"sync"
)

// Buffer tier sizes. A request is served from the smallest tier that fits it.
const (
	FourKB        uint32 = 4 << 10
	SixteenKB     uint32 = 16 << 10
	SixtyFourKB   uint32 = 64 << 10
	TwoFiftySixKB uint32 = 256 << 10
	OneMB         uint32 = 1 << 20
	FiveMB        uint32 = 5 << 20
	TenMB         uint32 = 10 << 20
)

// One pool per tier, initialized once by Preallocate.
var (
	pool4K, pool16K, pool64K, pool256K, pool1M, pool5M, pool10M sync.Pool
	preallocate                                                 sync.Once
)

// Preallocate initializes the tiered buffer pools. Must be called before SendFrame and ReceiveFrame.
func Preallocate() {
	preallocate.Do(internalAllocate)
}

func internalAllocate() {
	pool4K.New = newTier(FourKB)
	pool16K.New = newTier(SixteenKB)
	pool64K.New = newTier(SixtyFourKB)
	pool256K.New = newTier(TwoFiftySixKB)
	pool1M.New = newTier(OneMB)
	pool5M.New = newTier(FiveMB)
	pool10M.New = newTier(TenMB)
}

func newTier(size uint32) func() any {
	return func() any {
		data := make([]byte, size)
		return &data
	}
}

// get returns a buffer whose length is at least size. A size above the largest tier is allocated directly.
func get(size uint32) *[]byte {
	switch {
	case size <= FourKB:
		return pool4K.Get().(*[]byte)
	case size <= SixteenKB:
		return pool16K.Get().(*[]byte)
	case size <= SixtyFourKB:
		return pool64K.Get().(*[]byte)
	case size <= TwoFiftySixKB:
		return pool256K.Get().(*[]byte)
	case size <= OneMB:
		return pool1M.Get().(*[]byte)
	case size <= FiveMB:
		return pool5M.Get().(*[]byte)
	case size <= TenMB:
		return pool10M.Get().(*[]byte)
	default:
		data := make([]byte, size)
		return &data
	}
}

// put returns data to the tier its capacity came from. A buffer whose capacity is not a tier size,
// such as one allocated by get for a size above the largest tier, is dropped so it cannot
// be handed out for a request it does not fit or retained beyond its single use.
func put(data *[]byte) {
	switch uint32(cap(*data)) { //nolint:gosec // G115: capacity is bounded by the uint32 size given to get
	case FourKB:
		pool4K.Put(data)
	case SixteenKB:
		pool16K.Put(data)
	case SixtyFourKB:
		pool64K.Put(data)
	case TwoFiftySixKB:
		pool256K.Put(data)
	case OneMB:
		pool1M.Put(data)
	case FiveMB:
		pool5M.Put(data)
	case TenMB:
		pool10M.Put(data)
	}
}
