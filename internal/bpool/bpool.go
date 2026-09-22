// Package bpool is a tiered pool of byte buffers shared by the frame and the relay code.
// A request is served from the smallest tier that fits it; a buffer goes back only to the
// tier whose size equals its capacity, so nothing that was not taken from a tier is ever pooled.
package bpool

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

// tiers is indexed in parallel with tierSizes. The pools are initialized here, so the package
// needs no setup call before Get.
var tiers = [len(tierSizes)]sync.Pool{
	{New: newTier(FourKB)},
	{New: newTier(SixteenKB)},
	{New: newTier(SixtyFourKB)},
	{New: newTier(TwoFiftySixKB)},
	{New: newTier(OneMB)},
	{New: newTier(FiveMB)},
	{New: newTier(TenMB)},
}

func newTier(size uint32) func() any {
	return func() any {
		data := make([]byte, size)
		return &data
	}
}

// Get returns a buffer whose length is at least size. A size above the largest tier is allocated directly.
func Get(size uint32) *[]byte {
	for i, tier := range tierSizes {
		if size <= tier {
			return tiers[i].Get().(*[]byte)
		}
	}

	data := make([]byte, size)

	return &data
}

// Put returns data to the tier its capacity came from. A buffer whose capacity is not a tier size,
// such as one allocated by Get for a size above the largest tier, is dropped so it cannot
// be handed out for a request it does not fit or retained beyond its single use.
func Put(data *[]byte) {
	c := uint32(cap(*data)) //nolint:gosec // G115: capacity is bounded by the uint32 size given to Get
	for i, tier := range tierSizes {
		if c == tier {
			tiers[i].Put(data)
			return
		}
	}
}
