package goridge

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlloc(t *testing.T) {
	s := getAllocSize()
	if runtime.GOARCH == "amd64" {
		assert.Equal(t, s, uint(17179869184))
		d := make([]byte, s)
		_ = d
	} else if runtime.GOARCH == "386" {
		assert.Equal(t, s, uint(2147483648))
		d := make([]byte, s)
		_ = d
	}
}
