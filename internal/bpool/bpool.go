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

// One pool per tier, initialized here so the package needs no setup call before Get.
var (
	pool4K   = newTier(FourKB)
	pool16K  = newTier(SixteenKB)
	pool64K  = newTier(SixtyFourKB)
	pool256K = newTier(TwoFiftySixKB)
	pool1M   = newTier(OneMB)
	pool5M   = newTier(FiveMB)
	pool10M  = newTier(TenMB)
)

func newTier(size uint32) *sync.Pool {
	return &sync.Pool{
		New: func() any {
			data := make([]byte, size)
			return &data
		},
	}
}

// Get returns a buffer whose length is at least size. A size above the largest tier is allocated directly.
func Get(size int) *[]byte {
	switch {
	case size <= int(FourKB):
		return pool4K.Get().(*[]byte)
	case size <= int(SixteenKB):
		return pool16K.Get().(*[]byte)
	case size <= int(SixtyFourKB):
		return pool64K.Get().(*[]byte)
	case size <= int(TwoFiftySixKB):
		return pool256K.Get().(*[]byte)
	case size <= int(OneMB):
		return pool1M.Get().(*[]byte)
	case size <= int(FiveMB):
		return pool5M.Get().(*[]byte)
	case size <= int(TenMB):
		return pool10M.Get().(*[]byte)
	default:
		data := make([]byte, size)
		return &data
	}
}

// Put returns data to the tier its capacity came from. A buffer whose capacity is not a tier size,
// such as one allocated by Get for a size above the largest tier, is dropped so it cannot
// be handed out for a request it does not fit or retained beyond its single use.
func Put(data *[]byte) {
	switch cap(*data) {
	case int(FourKB):
		pool4K.Put(data)
	case int(SixteenKB):
		pool16K.Put(data)
	case int(SixtyFourKB):
		pool64K.Put(data)
	case int(TwoFiftySixKB):
		pool256K.Put(data)
	case int(OneMB):
		pool1M.Put(data)
	case int(FiveMB):
		pool5M.Put(data)
	case int(TenMB):
		pool10M.Put(data)
	}
}
