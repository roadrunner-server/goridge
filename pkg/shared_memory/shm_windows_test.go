// +build windows

package shared_memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testPayload string = "qwertyuioplkjhgfdsazxcvbnm" //26

func TestNewSharedMemory(t *testing.T) {
	shm, err := CreateSharedMemory("test", 1000)
	if err != nil {
		t.Fatal(shm)
	}

	d := make([]byte, 26, 26)

	for i := 0; i < 26; i++ {
		d[i] = testPayload[i]
	}

	shm.Write(d)

	d2 := make([]byte, 26)

	err = shm.Read(d2)
	if err != nil {
		t.Fatal(err)
	}

	err = shm.Detach()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, d, d2)
}

func TestSharedMemorySegment_Read(t *testing.T) {
	shm, err := CreateSharedMemory("test", 1000)
	if err != nil {
		t.Fatal(shm)
	}

	d2 := make([]byte, 26)

	err = shm.Read(d2)
	if err != nil {
		t.Fatal(err)
	}

}
