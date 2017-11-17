package goridge

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_min(t *testing.T) {
	assert.Equal(t, uint64(1), min(1, 2))
}
