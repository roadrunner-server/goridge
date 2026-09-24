package frame

import (
	"sync"
)

// Buffer tier sizes. A payload is served from the smallest tier that fits it with its headroom.
const (
	tier4K   = 4 << 10
	tier16K  = 16 << 10
	tier64K  = 64 << 10
	tier256K = 256 << 10
	tier1M   = 1 << 20
	tier5M   = 5 << 20
	tier10M  = 10 << 20
)

// One pool per tier, initialized here so that nothing needs a setup call.
var (
	pool4K   = newTier(tier4K)
	pool16K  = newTier(tier16K)
	pool64K  = newTier(tier64K)
	pool256K = newTier(tier256K)
	pool1M   = newTier(tier1M)
	pool5M   = newTier(tier5M)
	pool10M  = newTier(tier10M)
)

func newTier(size int) *sync.Pool {
	return &sync.Pool{
		New: func() any {
			data := make([]byte, size)
			return &data
		},
	}
}

// getBuf returns a buffer whose length is at least size. A size above the largest tier is allocated directly.
func getBuf(size int) *[]byte {
	switch {
	case size <= tier4K:
		return pool4K.Get().(*[]byte)
	case size <= tier16K:
		return pool16K.Get().(*[]byte)
	case size <= tier64K:
		return pool64K.Get().(*[]byte)
	case size <= tier256K:
		return pool256K.Get().(*[]byte)
	case size <= tier1M:
		return pool1M.Get().(*[]byte)
	case size <= tier5M:
		return pool5M.Get().(*[]byte)
	case size <= tier10M:
		return pool10M.Get().(*[]byte)
	default:
		data := make([]byte, size)
		return &data
	}
}

// putBuf returns data to the tier its capacity came from. A buffer whose capacity is not a tier size,
// such as one allocated by getBuf for a size above the largest tier, is dropped so it cannot
// be handed out for a request it does not fit or retained beyond its single use.
func putBuf(data *[]byte) {
	switch cap(*data) {
	case tier4K:
		pool4K.Put(data)
	case tier16K:
		pool16K.Put(data)
	case tier64K:
		pool64K.Put(data)
	case tier256K:
		pool256K.Put(data)
	case tier1M:
		pool1M.Put(data)
	case tier5M:
		pool5M.Put(data)
	case tier10M:
		pool10M.Put(data)
	}
}
