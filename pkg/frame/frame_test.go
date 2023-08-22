package frame

import (
	"hash/crc32"
	"testing"

	"github.com/stretchr/testify/assert"
)

const TestPayload = `alsdjf;lskjdgljasg;lkjsalfkjaskldjflkasjdf;lkasjfdalksdjflkajsdf;lfasdgnslsnblna;sldjjfawlkejr;lwjenlksndlfjawl;ejr;lwjelkrjaldfjl;sdjf`

func TestNewFrame(t *testing.T) {
	nf := NewFrame()
	nf.WriteVersion(nf.Header(), Version1)
	nf.WriteFlags(nf.Header(), CONTROL)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))
	nf.WriteCRC(nf.header)

	nf.WritePayload([]byte(TestPayload))

	data := nf.Bytes()

	rf := ReadFrame(data)

	assert.Equal(t, rf.ReadVersion(rf.Header()), nf.ReadVersion(nf.Header()))
	assert.Equal(t, rf.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, rf.ReadPayloadLen(rf.Header()), nf.ReadPayloadLen(nf.Header()))
	assert.Equal(t, true, rf.VerifyCRC(rf.Header()))
	assert.Equal(t, []uint32(nil), rf.ReadOptions(rf.Header()))
}

func TestAppendOptions(t *testing.T) {
	nf := NewFrame()
	nf.WriteVersion(nf.Header(), Version1)
	nf.WriteFlags(nf.Header(), CONTROL)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))
	nf.WriteCRC(nf.header)

	nf.AppendOptions(nf.HeaderPtr(), []byte{byte(112), byte(123), byte(0), byte(0)})

	nf.WritePayload([]byte(TestPayload))

	data := nf.Bytes()

	rf := ReadFrame(data)

	assert.Equal(t, rf.ReadVersion(rf.Header()), nf.ReadVersion(nf.Header()))
	assert.Equal(t, rf.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, rf.ReadPayloadLen(rf.Header()), nf.ReadPayloadLen(nf.Header()))
	assert.Equal(t, true, rf.VerifyCRC(rf.Header()))
	assert.Equal(t, []uint32(nil), rf.ReadOptions(rf.Header()))
}

func TestFrame_VerifyCRC_Fail(t *testing.T) {
	nf := NewFrame()
	// this is the wrong position
	nf.WriteCRC(nf.Header())
	nf.WriteVersion(nf.Header(), Version1)
	nf.WriteFlags(nf.Header(), CONTROL)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))

	nf.WritePayload([]byte(TestPayload))

	data := nf.Bytes()

	rf := ReadFrame(data)

	assert.Equal(t, rf.ReadVersion(rf.Header()), nf.ReadVersion(nf.Header()))
	assert.Equal(t, rf.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, rf.ReadPayloadLen(rf.Header()), nf.ReadPayloadLen(nf.Header()))
	assert.Equal(t, false, rf.VerifyCRC(rf.Header()))
}

func TestFrame_OptionsWithNoOptions(t *testing.T) {
	nf := NewFrame()
	nf.WriteVersion(nf.Header(), 1)
	nf.WriteFlags(nf.Header(), CONTROL, CodecGob)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))
	nf.WriteOptions(nf.HeaderPtr())

	// test options
	options := nf.ReadOptions(nf.Header())
	assert.Equal(t, []uint32(nil), options)
	// write payload
	nf.WritePayload([]byte(TestPayload))
	nf.WriteCRC(nf.Header())
	data := nf.Bytes()

	rf := ReadFrame(data)

	assert.Equal(t, rf.ReadVersion(rf.Header()), nf.ReadVersion(nf.Header()))
	assert.Equal(t, rf.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, rf.ReadPayloadLen(rf.Header()), nf.ReadPayloadLen(nf.Header()))
	assert.Equal(t, rf.VerifyCRC(rf.Header()), true)
}

func TestFrame_Panic(t *testing.T) {
	defer func() {
		assert.NotNil(t, recover())
	}()
	nf := NewFrame()
	nf.WriteVersion(nf.Header(), 1)
	nf.WriteFlags(nf.Header(), CONTROL, CodecGob)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))
	nf.WriteOptions(nf.HeaderPtr(), 323423432, 1213231, 1, 2, 3, 4, 5, 2, 1, 2, 12)
	nf.WriteOptions(nf.HeaderPtr(), 323423432)
}

func TestFrame_IncrementHLPanic(t *testing.T) {
	defer func() {
		assert.NotNil(t, recover())
	}()
	nf := NewFrame()
	nf.WriteVersion(nf.Header(), 1)
	nf.WriteFlags(nf.Header(), CONTROL, CodecGob)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))
	nf.WriteOptions(nf.HeaderPtr(), 323423432, 1213231, 1, 2, 3, 4, 5, 2, 1, 2)
	nf.incrementHL(nf.header)
	nf.incrementHL(nf.header)
	nf.incrementHL(nf.header)
}

func TestFrame_ReadOptionsPanic(t *testing.T) {
	defer func() {
		assert.NotNil(t, recover())
	}()
	nf := NewFrame()
	nf.WriteVersion(nf.Header(), 1)
	nf.WriteFlags(nf.Header(), CONTROL, CodecGob)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))
	nf.WriteOptions(nf.HeaderPtr(), 323423432, 1213231, 1, 2, 3, 4, 5, 2, 1, 2, 12)
	nf.header[53] = 123
}

func TestFrame_Options(t *testing.T) {
	nf := NewFrame()
	nf.WriteVersion(nf.Header(), 1)
	nf.WriteFlags(nf.Header(), CONTROL, CodecGob)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))
	nf.WriteOptions(nf.HeaderPtr(), 323423432, 1213231)

	// test options
	options := nf.ReadOptions(nf.Header())
	assert.Equal(t, []uint32{323423432, 1213231}, options)
	// write payload
	nf.WritePayload([]byte(TestPayload))
	nf.WriteCRC(nf.Header())
	data := nf.Bytes()

	rf := ReadFrame(data)

	assert.Equal(t, rf.ReadVersion(rf.Header()), nf.ReadVersion(nf.Header()))
	assert.Equal(t, rf.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, rf.ReadPayloadLen(rf.Header()), nf.ReadPayloadLen(nf.Header()))
	assert.Equal(t, rf.VerifyCRC(rf.Header()), true)
}

func TestFrame_Stream(t *testing.T) {
	nf := NewFrame()
	nf.WriteVersion(nf.Header(), 1)
	nf.WriteFlags(nf.Header(), CONTROL, CodecGob)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))
	nf.WriteOptions(nf.HeaderPtr(), 323423432, 1213231)
	nf.SetStreamFlag(nf.Header())

	// test options
	options := nf.ReadOptions(nf.Header())
	assert.Equal(t, []uint32{323423432, 1213231}, options)
	// write payload
	nf.WritePayload([]byte(TestPayload))
	nf.WriteCRC(nf.Header())
	data := nf.Bytes()

	rf := ReadFrame(data)

	assert.Equal(t, rf.ReadVersion(rf.Header()), nf.ReadVersion(nf.Header()))
	assert.Equal(t, rf.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, rf.ReadPayloadLen(rf.Header()), nf.ReadPayloadLen(nf.Header()))
	assert.Equal(t, true, rf.VerifyCRC(rf.Header()))
	assert.Equal(t, true, rf.IsStream(rf.Header()))
}

func TestFrame_Stop(t *testing.T) {
	nf := NewFrame()
	nf.WriteVersion(nf.Header(), 1)
	nf.WriteFlags(nf.Header(), CONTROL, CodecGob)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))
	nf.WriteOptions(nf.HeaderPtr(), 323423432, 1213231)
	nf.SetStreamFlag(nf.Header())
	nf.SetStopBit(nf.Header())

	// test options
	options := nf.ReadOptions(nf.Header())
	assert.Equal(t, []uint32{323423432, 1213231}, options)
	// write payload
	nf.WritePayload([]byte(TestPayload))
	nf.WriteCRC(nf.Header())
	data := nf.Bytes()

	rf := ReadFrame(data)

	assert.Equal(t, rf.ReadVersion(rf.Header()), nf.ReadVersion(nf.Header()))
	assert.Equal(t, rf.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, rf.ReadPayloadLen(rf.Header()), nf.ReadPayloadLen(nf.Header()))
	assert.Equal(t, true, rf.VerifyCRC(rf.Header()))
	assert.Equal(t, true, rf.IsStream(rf.Header()))
	assert.Equal(t, true, rf.IsStop(rf.Header()))
}

func TestFrame_Stop2(t *testing.T) {
	nf := NewFrame()
	nf.WriteVersion(nf.Header(), 1)
	nf.WriteFlags(nf.Header(), CONTROL, CodecGob)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))
	nf.WriteOptions(nf.HeaderPtr(), 323423432, 1213231)
	nf.SetStreamFlag(nf.Header())

	// test options
	options := nf.ReadOptions(nf.Header())
	assert.Equal(t, []uint32{323423432, 1213231}, options)
	// write payload
	nf.WritePayload([]byte(TestPayload))
	nf.WriteCRC(nf.Header())
	data := nf.Bytes()

	rf := ReadFrame(data)

	assert.Equal(t, rf.ReadVersion(rf.Header()), nf.ReadVersion(nf.Header()))
	assert.Equal(t, rf.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, rf.ReadPayloadLen(rf.Header()), nf.ReadPayloadLen(nf.Header()))
	assert.Equal(t, true, rf.VerifyCRC(rf.Header()))
	assert.Equal(t, true, rf.IsStream(rf.Header()))
	assert.Equal(t, false, rf.IsStop(rf.Header()))
}

func TestFrame_Stream2(t *testing.T) {
	nf := NewFrame()
	nf.WriteVersion(nf.Header(), 1)
	nf.WriteFlags(nf.Header(), CONTROL, CodecGob)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))
	nf.WriteOptions(nf.HeaderPtr(), 323423432, 1213231)

	// test options
	options := nf.ReadOptions(nf.Header())
	assert.Equal(t, []uint32{323423432, 1213231}, options)
	// write payload
	nf.WritePayload([]byte(TestPayload))
	nf.WriteCRC(nf.Header())
	data := nf.Bytes()

	rf := ReadFrame(data)

	assert.Equal(t, rf.ReadVersion(rf.Header()), nf.ReadVersion(nf.Header()))
	assert.Equal(t, rf.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, rf.ReadPayloadLen(rf.Header()), nf.ReadPayloadLen(nf.Header()))
	assert.Equal(t, true, rf.VerifyCRC(rf.Header()))
	assert.Equal(t, false, rf.IsStream(rf.Header()))
}

func BenchmarkLoops(b *testing.B) {
	nf := NewFrame()
	nf.WriteVersion(nf.Header(), 1)
	nf.WriteFlags(nf.Header(), CONTROL, CodecGob)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))
	nf.WriteOptions(nf.HeaderPtr(), 323423432, 1213231, 123123123, 398797979, 323423432, 1213231, 123123123, 398797979, 123, 123)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		options := nf.ReadOptions(nf.Header())
		_ = options
	}
}

func TestFrame_Bytes(t *testing.T) {
	nf := NewFrame()
	nf.WriteVersion(nf.Header(), 1)
	nf.WriteFlags(nf.Header(), CONTROL, CodecGob)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))

	nf.WriteOptions(nf.HeaderPtr(), 323423432)
	assert.Equal(t, []uint32{323423432}, nf.ReadOptions(nf.Header()))
	nf.WritePayload([]byte(TestPayload))

	nf.WriteCRC(nf.Header())
	assert.Equal(t, true, nf.VerifyCRC(nf.Header()))
	data := nf.Bytes()

	rf := ReadFrame(data)

	assert.Equal(t, rf.ReadVersion(rf.Header()), nf.ReadVersion(nf.Header()))
	assert.Equal(t, rf.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, rf.ReadPayloadLen(rf.Header()), nf.ReadPayloadLen(nf.Header()))
	assert.Equal(t, true, rf.VerifyCRC(rf.Header()))
	assert.Equal(t, []uint32{323423432}, rf.ReadOptions(rf.Header()))
}

func TestFrame_NotPingPong(t *testing.T) {
	nf := NewFrame()
	nf.WriteVersion(nf.Header(), 1)
	nf.WriteFlags(nf.Header(), CONTROL, CodecGob)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))

	nf.WriteOptions(nf.HeaderPtr(), 323423432)
	assert.Equal(t, []uint32{323423432}, nf.ReadOptions(nf.Header()))
	nf.WritePayload([]byte(TestPayload))

	nf.WriteCRC(nf.Header())
	assert.Equal(t, true, nf.VerifyCRC(nf.Header()))
	data := nf.Bytes()

	rf := ReadFrame(data)

	assert.False(t, rf.IsPing(rf.Header()))
	assert.False(t, rf.IsPong(rf.Header()))
	assert.Equal(t, rf.ReadVersion(rf.Header()), nf.ReadVersion(nf.Header()))
	assert.Equal(t, rf.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, rf.ReadPayloadLen(rf.Header()), nf.ReadPayloadLen(nf.Header()))
	assert.Equal(t, true, rf.VerifyCRC(rf.Header()))
	assert.Equal(t, []uint32{323423432}, rf.ReadOptions(rf.Header()))
}

func TestFrame_Ping(t *testing.T) {
	nf := NewFrame()
	nf.WriteVersion(nf.Header(), 1)
	nf.WriteFlags(nf.Header(), CONTROL, CodecGob)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))

	nf.WriteOptions(nf.HeaderPtr(), 323423432)
	assert.Equal(t, []uint32{323423432}, nf.ReadOptions(nf.Header()))
	nf.WritePayload([]byte(TestPayload))
	nf.SetPingBit(nf.Header())

	nf.WriteCRC(nf.Header())
	assert.Equal(t, true, nf.VerifyCRC(nf.Header()))
	data := nf.Bytes()

	rf := ReadFrame(data)

	assert.True(t, rf.IsPing(rf.Header()))
	assert.Equal(t, rf.ReadVersion(rf.Header()), nf.ReadVersion(nf.Header()))
	assert.Equal(t, rf.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, rf.ReadPayloadLen(rf.Header()), nf.ReadPayloadLen(nf.Header()))
	assert.Equal(t, true, rf.VerifyCRC(rf.Header()))
	assert.Equal(t, []uint32{323423432}, rf.ReadOptions(rf.Header()))
}

func TestFrame_Pong(t *testing.T) {
	nf := NewFrame()
	nf.WriteVersion(nf.Header(), 1)
	nf.WriteFlags(nf.Header(), CONTROL, CodecGob)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))

	nf.WriteOptions(nf.HeaderPtr(), 323423432)
	assert.Equal(t, []uint32{323423432}, nf.ReadOptions(nf.Header()))
	nf.WritePayload([]byte(TestPayload))
	nf.SetPongBit(nf.Header())

	nf.WriteCRC(nf.Header())
	assert.Equal(t, true, nf.VerifyCRC(nf.Header()))
	data := nf.Bytes()

	rf := ReadFrame(data)

	assert.True(t, rf.IsPong(rf.Header()))
	assert.Equal(t, rf.ReadVersion(rf.Header()), nf.ReadVersion(nf.Header()))
	assert.Equal(t, rf.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, rf.ReadPayloadLen(rf.Header()), nf.ReadPayloadLen(nf.Header()))
	assert.Equal(t, true, rf.VerifyCRC(rf.Header()))
	assert.Equal(t, []uint32{323423432}, rf.ReadOptions(rf.Header()))
}

func BenchmarkCRC32(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := crc32.ChecksumIEEE([]byte{'t', 't', 'b', 'u', '6', '1', 'g', 'h', 'r', 't'})
		_ = res
	}
}

func BenchmarkFrame_CRC(b *testing.B) {
	nf := NewFrame()
	nf.WriteVersion(nf.Header(), Version1)
	nf.WriteFlags(nf.Header(), CONTROL, CodecGob)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))
	nf.WriteOptions(nf.HeaderPtr(), 1000, 1000, 1000, 1000, 1000, 1000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		nf.WriteCRC(nf.Header())
		if !nf.VerifyCRC(nf.Header()) {
			panic("CRC")
		}
	}
}

func BenchmarkFrame(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		nf := NewFrame()
		nf.WriteVersion(nf.Header(), Version1)
		nf.WriteFlags(nf.Header(), CONTROL, CodecGob)
		nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))
		nf.WriteOptions(nf.HeaderPtr(), 1000, 1000, 1000, 1000, 1000, 1000)
		nf.WriteCRC(nf.Header())

		if !nf.VerifyCRC(nf.Header()) {
			panic("CRC")
		}
	}
}
