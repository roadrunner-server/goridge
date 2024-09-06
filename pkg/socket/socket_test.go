package socket

import (
	"net"
	"testing"

	"github.com/roadrunner-server/goridge/v3/pkg/frame"
	"github.com/stretchr/testify/assert"
)

const TestPayload = `alsdjf;lskjdgljasg;lkjsalfkjaskldjflkasjdf;lkasjfdalksdjflkajsdf;lfasdgnslsnblna;sldjjfawlkejr;lwjenlksndlfjawl;ejr;lwjelkrjaldfjl;sdjf`

func TestSocketRelay(t *testing.T) {
	// configure and create tcp4 listener
	ls, err := net.Listen("tcp", "localhost:10002")
	assert.NoError(t, err)

	// TEST FRAME TO SEND
	nf := frame.NewFrame()
	nf.WriteVersion(nf.Header(), frame.Version1)
	nf.WriteFlags(nf.Header(), frame.CONTROL, frame.CodecGob, frame.CodecJSON)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload)))) //nolint:gosec
	nf.WritePayload([]byte(TestPayload))
	nf.WriteCRC(nf.Header())
	assert.Equal(t, true, nf.VerifyCRC(nf.Header()))

	conn, err := net.Dial("tcp", "localhost:10002")
	assert.NoError(t, err)
	rsend := NewSocketRelay(conn)
	err = rsend.Send(nf)
	assert.NoError(t, err)

	accept, err := ls.Accept()
	assert.NoError(t, err)
	assert.NotNil(t, accept)

	r := NewSocketRelay(accept)

	fr := frame.NewFrame()
	err = r.Receive(fr)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fr.ReadVersion(fr.Header()), nf.ReadVersion(nf.Header()))
	assert.Equal(t, fr.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, fr.ReadPayloadLen(fr.Header()), nf.ReadPayloadLen(nf.Header()))
	assert.Equal(t, true, fr.VerifyCRC(fr.Header()))
	assert.Equal(t, []byte(TestPayload), fr.Payload())
}

func TestSocketRelayOptions(t *testing.T) {
	// configure and create tcp4 listener
	ls, err := net.Listen("tcp", "localhost:10001")
	assert.NoError(t, err)

	// TEST FRAME TO SEND
	nf := frame.NewFrame()
	nf.WriteVersion(nf.Header(), frame.Version1)
	nf.WriteFlags(nf.Header(), frame.CONTROL, frame.CodecGob, frame.CodecJSON)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload)))) //nolint:gosec
	nf.WritePayload([]byte(TestPayload))
	nf.WriteOptions(nf.HeaderPtr(), 100, 10000, 100000)
	nf.WriteCRC(nf.Header())
	assert.Equal(t, true, nf.VerifyCRC(nf.Header()))

	conn, err := net.Dial("tcp", "localhost:10001")
	assert.NoError(t, err)
	rsend := NewSocketRelay(conn)
	err = rsend.Send(nf)
	assert.NoError(t, err)

	accept, err := ls.Accept()
	assert.NoError(t, err)
	assert.NotNil(t, accept)

	r := NewSocketRelay(accept)

	fr := frame.NewFrame()
	err = r.Receive(fr)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fr.ReadVersion(fr.Header()), nf.ReadVersion(nf.Header()))
	assert.Equal(t, fr.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, fr.ReadPayloadLen(fr.Header()), nf.ReadPayloadLen(nf.Header()))
	assert.Equal(t, true, fr.VerifyCRC(fr.Header()))
	assert.Equal(t, []byte(TestPayload), fr.Payload())
	assert.Equal(t, []uint32{100, 10000, 100000}, fr.ReadOptions(fr.Header()))
}

func TestSocketRelayNoPayload(t *testing.T) {
	// configure and create tcp4 listener
	ls, err := net.Listen("tcp", "localhost:12221")
	assert.NoError(t, err)

	// TEST FRAME TO SEND
	nf := frame.NewFrame()
	nf.WriteVersion(nf.Header(), frame.Version1)
	nf.WriteFlags(nf.Header(), frame.CONTROL, frame.CodecGob, frame.CodecJSON)
	nf.WriteOptions(nf.HeaderPtr(), 100, 10000, 100000)
	nf.WriteCRC(nf.Header())
	assert.Equal(t, true, nf.VerifyCRC(nf.Header()))

	conn, err := net.Dial("tcp", "localhost:12221")
	assert.NoError(t, err)
	rsend := NewSocketRelay(conn)
	err = rsend.Send(nf)
	assert.NoError(t, err)

	accept, err := ls.Accept()
	assert.NoError(t, err)
	assert.NotNil(t, accept)

	r := NewSocketRelay(accept)

	fr := frame.NewFrame()
	err = r.Receive(fr)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fr.ReadVersion(fr.Header()), nf.ReadVersion(nf.Header()))
	assert.Equal(t, fr.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, fr.ReadPayloadLen(fr.Header()), nf.ReadPayloadLen(nf.Header()))
	assert.Equal(t, true, fr.VerifyCRC(fr.Header()))
	assert.Equal(t, []byte{}, fr.Payload()) // empty
	assert.Equal(t, []uint32{100, 10000, 100000}, fr.ReadOptions(fr.Header()))
}

func TestSocketRelayWrongCRC(t *testing.T) {
	// configure and create tcp4 listener
	ls, err := net.Listen("tcp", "localhost:13445")
	assert.NoError(t, err)

	// TEST FRAME TO SEND
	nf := frame.NewFrame()
	nf.WriteVersion(nf.Header(), frame.Version1)
	nf.WriteFlags(nf.Header(), frame.CONTROL, frame.CodecGob, frame.CodecJSON)
	nf.WriteOptions(nf.HeaderPtr(), 100, 10000, 100000)
	nf.WriteCRC(nf.Header())
	nf.Header()[6] = 22 // just random wrong CRC directly

	conn, err := net.Dial("tcp", "localhost:13445")
	assert.NoError(t, err)
	_, err = conn.Write(nf.Bytes())
	assert.NoError(t, err)

	accept, err := ls.Accept()
	assert.NoError(t, err)
	assert.NotNil(t, accept)

	r := NewSocketRelay(accept)

	fr := frame.NewFrame()
	err = r.Receive(fr)
	assert.Error(t, err)
	assert.False(t, fr.VerifyCRC(fr.Header()))

	assert.Empty(t, fr.Payload())
}
