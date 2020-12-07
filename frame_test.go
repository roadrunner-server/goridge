package goridge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const TestPayload = `alsdjf;lskjdgljasg;lkjsalfkjaskldjflkasjdf;lkasjfdalksdjflkajsdf;lfasdgnslsnblna;sldjjfawlkejr;lwjenlksndlfjawl;ejr;lwjelkrjaldfjl;sdjf`

func TestNewFrame(t *testing.T) {
	initLookupTable()

	nf := NewFrame()
	nf.WriteVersion(1)
	nf.WriteHL(5)
	nf.WriteFlags(12)
	nf.WritePayloadLen(uint32(len([]byte(TestPayload))))
	nf.WriteCRC()

	nf.WritePayload([]byte(TestPayload))

	data := nf.Bytes()

	rf := ReadFrame(data)

	assert.Equal(t, rf.ReadVersion(), nf.ReadVersion())
	assert.Equal(t, rf.ReadHL(), nf.ReadHL())
	assert.Equal(t, rf.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, rf.ReadPayloadLen(), nf.ReadPayloadLen())
	assert.Equal(t, rf.VerifyCRC(), true)
}

func BenchmarkFrame_Bytes(b *testing.B) {
	initLookupTable()
	nf := NewFrame()
	nf.WriteVersion(1)
	nf.WriteHL(5)
	nf.WriteFlags(12)
	nf.WritePayloadLen(uint32(len([]byte(TestPayload))))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		nf.WriteCRC()
		if !nf.VerifyCRC() {
			panic("CRC")
		}
	}
}
