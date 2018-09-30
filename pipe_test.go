package goridge

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClosePipeRelay(t *testing.T) {
	r := NewPipeRelay(&connMock{}, &connMock{})
	assert.Nil(t, r.Close())
}

func TestPipeReceive(t *testing.T) {
	conn := &connMock{}
	r := NewPipeRelay(conn, &connMock{})
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

func TestPipeSend(t *testing.T) {
	conn := &connMock{}
	r := NewPipeRelay(&connMock{}, conn)
	assert.Nil(t, r.Close())

	prefix := NewPrefix().WithFlag(PayloadControl).WithSize(5)
	payload := []byte("hello")

	conn.expect(write, append(prefix[:], payload...))

	err := r.Send(payload, prefix.Flags())
	assert.Nil(t, err)
	assert.Empty(t, 0, conn.leftSegments())
}
