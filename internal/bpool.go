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

// tierSizes lists the tiers in ascending order. Every pooled buffer has exactly one of these capacities.
var tierSizes = [...]uint32{FourKB, SixteenKB, SixtyFourKB, TwoFiftySixKB, OneMB, FiveMB, TenMB}

var (
	tiers       [len(tierSizes)]sync.Pool
	preallocate sync.Once
)

// Preallocate initializes the tiered buffer pools. Must be called before SendFrame and ReceiveFrame.
func Preallocate() {
	preallocate.Do(internalAllocate)
}

func internalAllocate() {
	for i, size := range tierSizes {
		tiers[i].New = func() any {
			data := make([]byte, size)
			return &data
		}
	}
}

// get returns a buffer whose length is at least size. A size above the largest tier is allocated directly.
func get(size uint32) *[]byte {
	for i, tier := range tierSizes {
		if size <= tier {
			return tiers[i].Get().(*[]byte)
		}
	}

	data := make([]byte, size)

	return &data
}

// put returns data to the tier its capacity came from. A buffer whose capacity is not a tier size,
// such as one allocated by get for a size above the largest tier, is dropped so it cannot
// be handed out for a request it does not fit or retained beyond its single use.
func put(data *[]byte) {
	c := uint32(cap(*data)) //nolint:gosec // G115: capacity is bounded by the uint32 size given to get
	for i, tier := range tierSizes {
		if c == tier {
			tiers[i].Put(data)
			return
		}
	}
}
