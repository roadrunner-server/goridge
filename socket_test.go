package goridge

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCloseSocketRelay(t *testing.T) {
	m := &connMock{}
	r := NewSocketRelay(m)

	assert.False(t, m.closed)
	r.Close()
	assert.True(t, m.closed)
}

func TestSocketReceive(t *testing.T) {
	conn := &connMock{}
	r := NewSocketRelay(conn)
	assert.Nil(t, r.Close())

	prefix := NewPrefix().WithFlag(PayloadControl).WithSize(5)
	payload := []byte("hello")

	conn.expect(read, prefix[:])
	conn.expect(read, payload)

	data, p, err := r.Receive()

	assert.Nil(t, err)
	assert.True(t, p.HasFlag(PayloadControl))
	assert.Equal(t, uint64(5), p.Size())
	assert.Equal(t, 0, bytes.Compare(data, payload))
	assert.Empty(t, 0, conn.leftSegments())
}

func TestSocketSend(t *testing.T) {
	conn := &connMock{}
	r := NewSocketRelay(conn)
	assert.Nil(t, r.Close())

	prefix := NewPrefix().WithFlag(PayloadControl).WithSize(5)
	payload := []byte("hello")

	conn.expect(write, prefix[:])
	conn.expect(write, payload)

	err := r.Send(payload, prefix.Flags())
	assert.Nil(t, err)
	assert.Empty(t, 0, conn.leftSegments())
}
