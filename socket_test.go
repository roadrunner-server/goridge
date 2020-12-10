package goridge

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSocketRelay(t *testing.T) {
	// configure and create tcp4 listener
	ls, err := net.Listen("tcp", "localhost:10002")
	assert.NoError(t, err)

	// TEST FRAME TO SEND
	nf := NewFrame()
	nf.WriteVersion(VERSION_1)
	nf.WriteFlags(CONTROL, CODEC_GOB, CODEC_JSON)
	nf.WritePayloadLen(uint32(len([]byte(TestPayload))))
	nf.WritePayload([]byte(TestPayload))
	nf.WriteCRC()
	assert.Equal(t, true, nf.VerifyCRC())

	conn, err := net.Dial("tcp", "localhost:10002")
	assert.NoError(t, err)
	rsend := NewSocketRelay(conn)
	err = rsend.Send(nf)
	assert.NoError(t, err)

	accept, err := ls.Accept()
	assert.NoError(t, err)
	assert.NotNil(t, accept)

	r := NewSocketRelay(accept)

	frame := &Frame{}
	err = r.Receive(frame)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, frame.ReadVersion(), nf.ReadVersion())
	assert.Equal(t, frame.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, frame.ReadPayloadLen(), nf.ReadPayloadLen())
	assert.Equal(t, true, frame.VerifyCRC())
	assert.Equal(t, []byte(TestPayload), frame.Payload())
}

func TestSocketRelayOptions(t *testing.T) {
	// configure and create tcp4 listener
	ls, err := net.Listen("tcp", "localhost:10001")
	assert.NoError(t, err)

	// TEST FRAME TO SEND
	nf := NewFrame()
	nf.WriteVersion(VERSION_1)
	nf.WriteFlags(CONTROL, CODEC_GOB, CODEC_JSON)
	nf.WritePayloadLen(uint32(len([]byte(TestPayload))))
	nf.WritePayload([]byte(TestPayload))
	nf.WriteOptions(100, 10000, 100000)
	nf.WriteCRC()
	assert.Equal(t, true, nf.VerifyCRC())

	conn, err := net.Dial("tcp", "localhost:10001")
	assert.NoError(t, err)
	rsend := NewSocketRelay(conn)
	err = rsend.Send(nf)
	assert.NoError(t, err)

	accept, err := ls.Accept()
	assert.NoError(t, err)
	assert.NotNil(t, accept)

	r := NewSocketRelay(accept)

	frame := &Frame{}
	err = r.Receive(frame)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, frame.ReadVersion(), nf.ReadVersion())
	assert.Equal(t, frame.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, frame.ReadPayloadLen(), nf.ReadPayloadLen())
	assert.Equal(t, true, frame.VerifyCRC())
	assert.Equal(t, []byte(TestPayload), frame.Payload())
	assert.Equal(t, []uint32{100, 10000, 100000}, frame.ReadOptions())
}

func TestSocketRelayNoPayload(t *testing.T) {
	// configure and create tcp4 listener
	ls, err := net.Listen("tcp", "localhost:12221")
	assert.NoError(t, err)

	// TEST FRAME TO SEND
	nf := NewFrame()
	nf.WriteVersion(VERSION_1)
	nf.WriteFlags(CONTROL, CODEC_GOB, CODEC_JSON)
	nf.WriteOptions(100, 10000, 100000)
	nf.WriteCRC()
	assert.Equal(t, true, nf.VerifyCRC())

	conn, err := net.Dial("tcp", "localhost:12221")
	assert.NoError(t, err)
	rsend := NewSocketRelay(conn)
	err = rsend.Send(nf)
	assert.NoError(t, err)

	accept, err := ls.Accept()
	assert.NoError(t, err)
	assert.NotNil(t, accept)

	r := NewSocketRelay(accept)

	frame := &Frame{}
	err = r.Receive(frame)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, frame.ReadVersion(), nf.ReadVersion())
	assert.Equal(t, frame.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, frame.ReadPayloadLen(), nf.ReadPayloadLen()) // should be zero, without error
	assert.Equal(t, true, frame.VerifyCRC())
	assert.Equal(t, []byte{}, frame.Payload()) // empty
	assert.Equal(t, []uint32{100, 10000, 100000}, frame.ReadOptions())
}

func TestSocketRelayWrongCRC(t *testing.T) {
	// configure and create tcp4 listener
	ls, err := net.Listen("tcp", "localhost:13445")
	assert.NoError(t, err)

	// TEST FRAME TO SEND
	nf := NewFrame()
	nf.WriteVersion(VERSION_1)
	nf.WriteFlags(CONTROL, CODEC_GOB, CODEC_JSON)
	nf.WriteOptions(100, 10000, 100000)
	nf.WriteCRC()
	nf.header[6] = 22 // just random wrong CRC directly

	conn, err := net.Dial("tcp", "localhost:13445")
	assert.NoError(t, err)
	_, err = conn.Write(nf.Bytes())
	assert.NoError(t, err)

	accept, err := ls.Accept()
	assert.NoError(t, err)
	assert.NotNil(t, accept)

	r := NewSocketRelay(accept)

	frame := &Frame{}
	err = r.Receive(frame)
	assert.Error(t, err)
	assert.Nil(t, frame.header)
	assert.Nil(t, frame.payload)
}
