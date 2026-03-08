package frame

import (
	"testing"
)

func FuzzReadFrame(f *testing.F) {
	// Seed: valid frame with payload
	nf := NewFrame()
	nf.WriteVersion(nf.Header(), Version1)
	nf.WriteFlags(nf.Header(), CONTROL)
	nf.WritePayloadLen(nf.Header(), 5)
	nf.WriteCRC(nf.Header())
	nf.WritePayload([]byte("hello"))
	f.Add(nf.Bytes())

	// Seed: frame with options
	nf2 := NewFrame()
	nf2.WriteVersion(nf2.Header(), Version1)
	nf2.WriteFlags(nf2.Header(), CodecJSON)
	nf2.WriteOptions(nf2.HeaderPtr(), 42, 10)
	nf2.WritePayloadLen(nf2.Header(), 3)
	nf2.WriteCRC(nf2.Header())
	nf2.WritePayload([]byte("abc"))
	f.Add(nf2.Bytes())

	// Seed: 12 zero bytes (minimal)
	f.Add(make([]byte, 12))

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() { recover() }() //nolint:errcheck

		fr := ReadFrame(data)
		// Exercise all read methods on the result
		_ = fr.ReadVersion(fr.Header())
		_ = fr.ReadFlags()
		_ = fr.ReadHL(fr.Header())
		_ = fr.ReadPayloadLen(fr.Header())
		_ = fr.VerifyCRC(fr.Header())
		_ = fr.ReadOptions(fr.Header())
		_ = fr.Payload()
		_ = fr.Header()
	})
}

func FuzzReadOptions(f *testing.F) {
	// Seed: hl=3 (12-byte header, no options)
	h1 := make([]byte, 12)
	h1[0] = 3
	f.Add(h1)

	// Seed: hl=4 (16-byte header, 1 option)
	h2 := make([]byte, 16)
	h2[0] = 4
	f.Add(h2)

	// Seed: hl=13 (52-byte header, 10 options — max)
	h3 := make([]byte, 52)
	h3[0] = 13
	f.Add(h3)

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() { recover() }() //nolint:errcheck

		fr := NewFrame()
		_ = fr.ReadOptions(data)
	})
}

func FuzzVerifyCRC(f *testing.F) {
	// Seed: valid CRC header
	nf := NewFrame()
	nf.WriteVersion(nf.Header(), Version1)
	nf.WriteFlags(nf.Header(), CONTROL)
	nf.WritePayloadLen(nf.Header(), 0)
	nf.WriteCRC(nf.Header())
	f.Add(nf.Header())

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 10 {
			return // VerifyCRC requires at least 10 bytes
		}
		defer func() { recover() }() //nolint:errcheck

		fr := NewFrame()
		_ = fr.VerifyCRC(data)
	})
}

func FuzzFrameRoundTrip(f *testing.F) {
	// Structured seeds: version, flags, payloadLen as uint32, then payload bytes
	f.Add(byte(1), byte(0x08), uint32(5), []byte("hello"))
	f.Add(byte(0), byte(0x20), uint32(0), []byte{})
	f.Add(byte(15), byte(0x04), uint32(3), []byte{0xFF, 0xFE, 0xFD})

	f.Fuzz(func(t *testing.T, version byte, flags byte, payloadLen uint32, payload []byte) {
		defer func() { recover() }() //nolint:errcheck

		// Cap version to valid range
		if version > 15 {
			version %= 16
		}

		// Cap payload to 1MB to avoid OOM
		const maxPayload = 1 << 20
		if len(payload) > maxPayload {
			payload = payload[:maxPayload]
		}

		nf := NewFrame()
		nf.WriteVersion(nf.Header(), version)
		nf.WriteFlags(nf.Header(), flags)

		// Use actual payload length, not the fuzzed payloadLen
		nf.WritePayloadLen(nf.Header(), uint32(len(payload))) //nolint:gosec
		nf.WritePayload(payload)
		nf.WriteCRC(nf.Header())

		data := nf.Bytes()
		rf := ReadFrame(data)

		if rf.ReadVersion(rf.Header()) != version {
			t.Errorf("version mismatch: got %d, want %d", rf.ReadVersion(rf.Header()), version)
		}
		if rf.ReadPayloadLen(rf.Header()) != uint32(len(payload)) { //nolint:gosec
			t.Errorf("payload length mismatch: got %d, want %d", rf.ReadPayloadLen(rf.Header()), len(payload))
		}
		if !rf.VerifyCRC(rf.Header()) {
			t.Error("CRC verification failed on round-tripped frame")
		}
	})
}
