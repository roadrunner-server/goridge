package goridge

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPrefix(t *testing.T) {
	p1 := NewPrefix()

	assert.Equal(t, byte(0), p1.Flags())
	assert.Equal(t, uint64(0), p1.Size())
}

func TestPrefix_WithFlag(t *testing.T) {
	p1 := NewPrefix()
	p2 := p1.WithFlag(PayloadRaw)

	assert.False(t, p1.HasFlag(PayloadRaw))
	assert.True(t, p2.HasFlag(PayloadRaw))

	p3 := p2.WithFlag(PayloadEmpty)
	assert.True(t, p3.HasFlag(PayloadRaw))
	assert.True(t, p3.HasFlag(PayloadEmpty))
}

func TestPrefix_WithFlags(t *testing.T) {
	p1 := NewPrefix().WithFlag(PayloadRaw)
	p2 := p1.WithFlags(PayloadEmpty)

	assert.False(t, p2.HasFlag(PayloadRaw))
	assert.True(t, p2.HasFlag(PayloadEmpty))
}

func TestPrefix_WithSize(t *testing.T) {
	p1 := NewPrefix().WithFlag(PayloadRaw)
	p2 := p1.WithSize(1000)

	assert.True(t, p1.HasFlag(PayloadRaw))
	assert.Equal(t, uint64(0), p1.Size())

	assert.True(t, p2.HasFlag(PayloadRaw))
	assert.Equal(t, uint64(1000), p2.Size())
}

func TestPrefix_HasPayload(t *testing.T) {
	p1 := NewPrefix().WithFlag(PayloadRaw)
	p2 := p1.WithSize(1000)

	assert.False(t, p1.HasPayload())
	assert.True(t, p2.HasPayload())

	p3 := p2.WithFlag(PayloadEmpty)
	assert.False(t, p3.HasPayload())
}

func TestReadPrefix(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{PayloadRaw | PayloadControl, 255, 0, 0, 0, 255, 0, 0, 0, 0, 0, 0, 255, 0, 0, 0, 255})

	p1 := NewPrefix()
	_, err := buffer.Read(p1[:])
	if err != nil {
		t.Errorf("error during reading the buffer: %v", err)
	}

	assert.True(t, p1.HasFlag(PayloadRaw))
	assert.True(t, p1.HasFlag(PayloadControl))

	assert.Equal(t, uint64(1095216660735), p1.Size())

	assert.True(t, p1.Valid())
}

func TestInvalidPrefix(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{PayloadRaw | PayloadControl, 255, 0, 0, 0, 255, 0, 0, 0, 0, 0, 0, 255, 0, 0, 0, 254})

	p1 := NewPrefix()
	_, err := buffer.Read(p1[:])
	if err != nil {
		t.Errorf("error during reading the buffer: %v", err)
	}

	assert.False(t, p1.Valid())
}
