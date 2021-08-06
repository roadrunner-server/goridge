package pipe

import (
	"io"
	"testing"

	"github.com/spiral/goridge/v3/pkg/frame"
	"github.com/stretchr/testify/assert"
)

const TestPayload = `alsdjf;lskjdgljasg;lkjsalfkjaskldjflkasjdf;lkasjfdalksdjflkajsdf;lfasdgnslsnblna;sldjjfawlkejr;lwjenlksndlfjawl;ejr;lwjelkrjaldfjl;sdjf`

func TestPipeReceive(t *testing.T) {
	pr, pw := io.Pipe()

	relay := NewPipeRelay(pr, pw)

	nf := frame.NewFrame()
	nf.WriteVersion(nf.Header(), frame.VERSION_1)
	nf.WriteFlags(frame.CONTROL, frame.CODEC_GOB, frame.CODEC_JSON)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))
	nf.WritePayload([]byte(TestPayload))
	nf.WriteCRC(nf.Header())
	assert.Equal(t, true, nf.VerifyCRC(nf.Header()))

	go func(frame *frame.Frame) {
		defer func() {
			_ = pw.Close()
		}()
		err := relay.Send(nf)
		assert.NoError(t, err)
		_ = pw.Close()
	}(nf)

	fr := frame.NewFrame()
	err := relay.Receive(fr)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fr.ReadVersion(fr.Header()), nf.ReadVersion(nf.Header()))
	assert.Equal(t, fr.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, fr.ReadPayloadLen(fr.Header()), nf.ReadPayloadLen(nf.Header()))
	assert.Equal(t, true, fr.VerifyCRC(nf.Header()))
	assert.Equal(t, []byte(TestPayload), fr.Payload())
}

func TestPipeReceiveWithOptions(t *testing.T) {
	pr, pw := io.Pipe()

	relay := NewPipeRelay(pr, pw)

	nf := frame.NewFrame()
	nf.WriteVersion(nf.Header(), frame.VERSION_1)
	nf.WriteFlags(frame.CONTROL, frame.CODEC_GOB, frame.CODEC_JSON)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))
	nf.WritePayload([]byte(TestPayload))
	nf.WriteOptions(nf.HeaderPtr(), 100, 10000, 100000)
	nf.WriteCRC(nf.Header())
	assert.Equal(t, true, nf.VerifyCRC(nf.Header()))

	go func(frame *frame.Frame) {
		defer func() {
			_ = pw.Close()
		}()
		err := relay.Send(nf)
		assert.NoError(t, err)
		_ = pw.Close()
	}(nf)

	fr := frame.NewFrame()
	err := relay.Receive(fr)
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

func TestPipeCRC_Failed(t *testing.T) {
	pr, pw := io.Pipe()

	relay := NewPipeRelay(pr, pw)

	nf := frame.NewFrame()
	nf.WriteVersion(nf.Header(), frame.VERSION_1)
	nf.WriteFlags(frame.CONTROL)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))

	assert.Equal(t, false, nf.VerifyCRC(nf.Header()))

	nf.WritePayload([]byte(TestPayload))

	go func(frame *frame.Frame) {
		defer func() {
			_ = pw.Close()
		}()
		err := relay.Send(nf)
		assert.NoError(t, err)
		_ = pw.Close()
	}(nf)

	fr := frame.NewFrame()
	err := relay.Receive(fr)
	assert.Error(t, err)
	assert.False(t, fr.VerifyCRC(fr.Header()))

	assert.Empty(t, fr.Payload())
}
