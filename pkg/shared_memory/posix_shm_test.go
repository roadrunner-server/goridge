// +build linux

package shared_memory //nolint:golint,stylecheck

import (
	"testing"

	"github.com/spiral/goridge/v3/pkg/shared_memory/test"
	"github.com/stretchr/testify/assert"
)

const testData = "hello my dear friend"

func TestNewSharedMemorySegment(t *testing.T) {
	testBuf := make([]byte, 0)
	testBuf = append(testBuf, []byte(testData)...)

	seg1, err := NewSharedMemorySegment(0x1, 1024, S_IRUSR|S_IWUSR|S_IRGRP|S_IWGRP, IPC_CREAT)
	assert.NoError(t, err)

	// write data to the shared memory
	seg1.Write([]byte(testData))
	err = seg1.Detach()
	assert.NoError(t, err)

	seg2, err := NewSharedMemorySegment(0x1, 1024, 0, SHM_RDONLY)
	assert.NoError(t, err)

	buf := make([]byte, len(testData))
	err = seg2.Read(buf)
	assert.NoError(t, err)

	err = seg2.Detach()
	assert.NoError(t, err)

	assert.Equal(t, testBuf, buf)
}

func TestAttachToShmSegment(t *testing.T) {
	testBuf := make([]byte, 0)
	testBuf = append(testBuf, []byte(testData)...)
	// Just to be sure, that shm segment exists
	seg1, err := NewSharedMemorySegment(0x1, 1024, S_IRUSR|S_IWUSR|S_IRGRP|S_IWGRP, IPC_CREAT)
	assert.NoError(t, err)

	// clear shm segment
	seg1.Clear()

	// write data to the shared memory
	seg1.Write([]byte(testData))
	err = seg1.Detach()
	assert.NoError(t, err)

	seg2, err := AttachToShmSegment(int(seg1.address), 1024, 0666)
	assert.NoError(t, err)

	buf := make([]byte, len(testData))
	err = seg2.Read(buf)
	assert.NoError(t, err)

	err = seg2.Detach()
	assert.NoError(t, err)

	assert.Equal(t, testBuf, buf)
}

// 75 microseconds - Read
func BenchmarkAttachToShmSegment_READ(b *testing.B) {
	bigJSONLen := len(test.BigJSON)
	testBuf := make([]byte, 0, len(testData))
	testBuf = append(testBuf, testData...)
	// Just to be sure, that shm segment exists
	seg1, err := NewSharedMemorySegment(0x10, uint(bigJSONLen), S_IRUSR|S_IWUSR|S_IRGRP|S_IWGRP, IPC_CREAT)
	assert.NoError(b, err)

	// clear shm segment
	seg1.Clear()

	// write data to the shared memory
	seg1.Write(testBuf)
	err = seg1.Detach()
	assert.NoError(b, err)

	seg2, err := AttachToShmSegment(int(seg1.address), uint(bigJSONLen), 0666)
	assert.NoError(b, err)

	buf := make([]byte, bigJSONLen)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = seg2.Read(buf)
		if err != nil {
			b.Fatal(err)
		}
	}

	err = seg2.Detach()
	assert.NoError(b, err)
}

// 135 microseconds - Write
// 50880	     23679 ns/op	  147456 B/op	       1 allocs/op
// 10639	    152172 ns/op	  147456 B/op	       1 allocs/op
func BenchmarkAttachToShmSegment_WRITE(b *testing.B) {
	bigJSONLen := len(test.BigJSON)
	testBuf := make([]byte, 0, len(testData))
	testBuf = append(testBuf, testData...)
	// Just to be sure, that shm segment exists
	seg1, err := NewSharedMemorySegment(0x20, uint(bigJSONLen), S_IRUSR|S_IWUSR|S_IRGRP|S_IWGRP, IPC_CREAT)
	if err != nil {
		b.Fatal(err)
	}

	// clear shm segment
	seg1.Clear()

	// write data to the shared memory
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		seg1.Write(testBuf)
		seg1.Clear()
	}

	err = seg1.Detach()
	if err != nil {
		b.Fatal(err)
	}
}
