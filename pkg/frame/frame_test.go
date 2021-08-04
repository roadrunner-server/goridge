package frame

import (
	"hash/crc32"
	"testing"

	"github.com/stretchr/testify/assert"
)

const TestPayload = `alsdjf;lskjdgljasg;lkjsalfkjaskldjflkasjdf;lkasjfdalksdjflkajsdf;lfasdgnslsnblna;sldjjfawlkejr;lwjenlksndlfjawl;ejr;lwjelkrjaldfjl;sdjf`

func TestNewFrame(t *testing.T) {
	nf := NewFrame()
	nf.WriteVersion(VERSION_1)
	nf.WriteFlags(CONTROL)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))
	nf.WriteCRC(nf.header)

	nf.WritePayload([]byte(TestPayload))

	data := nf.Bytes()

	rf := ReadFrame(data)

	assert.Equal(t, rf.ReadVersion(), nf.ReadVersion())
	assert.Equal(t, rf.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, rf.ReadPayloadLen(), nf.ReadPayloadLen())
	assert.Equal(t, true, rf.VerifyCRC(rf.Header()))
	assert.Equal(t, []uint32(nil), rf.ReadOptions(rf.Header()))
}

func TestFrame_VerifyCRC_Fail(t *testing.T) {
	nf := NewFrame()
	// this is the wrong position
	nf.WriteCRC(nf.Header())
	nf.WriteVersion(VERSION_1)
	nf.WriteFlags(CONTROL)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))

	nf.WritePayload([]byte(TestPayload))

	data := nf.Bytes()

	rf := ReadFrame(data)

	assert.Equal(t, rf.ReadVersion(), nf.ReadVersion())
	assert.Equal(t, rf.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, rf.ReadPayloadLen(), nf.ReadPayloadLen())
	assert.Equal(t, false, rf.VerifyCRC(rf.Header()))
}

func TestFrame_Options(t *testing.T) {
	nf := NewFrame()
	nf.WriteVersion(1)
	nf.WriteFlags(CONTROL, CODEC_GOB)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))
	nf.WriteOptions(323423432, 1213231)

	// test options
	options := nf.ReadOptions(nf.Header())
	assert.Equal(t, []uint32{323423432, 1213231}, options)
	// write payload
	nf.WritePayload([]byte(TestPayload))
	nf.WriteCRC(nf.Header())
	data := nf.Bytes()

	rf := ReadFrame(data)

	assert.Equal(t, rf.ReadVersion(), nf.ReadVersion())
	assert.Equal(t, rf.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, rf.ReadPayloadLen(), nf.ReadPayloadLen())
	assert.Equal(t, rf.VerifyCRC(rf.Header()), true)
}

func TestFrame_Bytes(t *testing.T) {
	nf := NewFrame()
	nf.WriteVersion(1)
	nf.WriteFlags(CONTROL, CODEC_GOB)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))

	nf.WriteOptions(323423432)
	assert.Equal(t, []uint32{323423432}, nf.ReadOptions(nf.Header()))
	nf.WritePayload([]byte(TestPayload))

	nf.WriteCRC(nf.Header())
	assert.Equal(t, true, nf.VerifyCRC(nf.Header()))
	data := nf.Bytes()

	rf := ReadFrame(data)

	assert.Equal(t, rf.ReadVersion(), nf.ReadVersion())
	assert.Equal(t, rf.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, rf.ReadPayloadLen(), nf.ReadPayloadLen())
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
	nf.WriteVersion(VERSION_1)
	nf.WriteFlags(CONTROL, CODEC_GOB)
	nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))
	nf.WriteOptions(1000, 1000, 1000, 1000, 1000, 1000)

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
		nf.WriteVersion(VERSION_1)
		nf.WriteFlags(CONTROL, CODEC_GOB)
		nf.WritePayloadLen(nf.Header(), uint32(len([]byte(TestPayload))))
		nf.WriteOptions(1000, 1000, 1000, 1000, 1000, 1000)
		nf.WriteCRC(nf.Header())

		if !nf.VerifyCRC(nf.Header()) {
			panic("CRC")
		}
	}
}
