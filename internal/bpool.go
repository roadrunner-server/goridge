package internal

import (
	"sync"
)

var frameChunkedPool = &sync.Map{}
var preallocate = &sync.Once{}

const (
	OneMB  uint32 = 1024 * 1024 * 1
	FiveMB uint32 = 1024 * 1024 * 5
	TenMB  uint32 = 1024 * 1024 * 10
)

func Preallocate() {
	preallocate.Do(internalAllocate)
}

func internalAllocate() {
	pool1 := &sync.Pool{
		New: func() any {
			data := make([]byte, OneMB)
			return &data
		},
	}
	pool5 := &sync.Pool{
		New: func() any {
			data := make([]byte, FiveMB)
			return &data
		},
	}
	pool10 := &sync.Pool{
		New: func() any {
			data := make([]byte, TenMB)
			return &data
		},
	}

	frameChunkedPool.Store(OneMB, pool1)
	frameChunkedPool.Store(FiveMB, pool5)
	frameChunkedPool.Store(TenMB, pool10)
}

func get(size uint32) *[]byte {
	switch {
	case size <= OneMB:
		val, _ := frameChunkedPool.Load(OneMB)
		return val.(*sync.Pool).Get().(*[]byte)
	case size <= FiveMB:
		val, _ := frameChunkedPool.Load(FiveMB)
		return val.(*sync.Pool).Get().(*[]byte)
	case size <= TenMB:
		val, _ := frameChunkedPool.Load(TenMB)
		return val.(*sync.Pool).Get().(*[]byte)
	default:
		data := make([]byte, size)
		return &data
	}
}

func put(size uint32, data *[]byte) {
	switch {
	case size <= OneMB:
		pool, _ := frameChunkedPool.Load(OneMB)
		pool.(*sync.Pool).Put(data)
		return
	case size <= FiveMB:
		pool, _ := frameChunkedPool.Load(FiveMB)
		pool.(*sync.Pool).Put(data)
		return
	default:
		pool, _ := frameChunkedPool.Load(TenMB)
		pool.(*sync.Pool).Put(data)
		return
	}
}
