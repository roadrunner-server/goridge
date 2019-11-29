package goridge

import (
	"encoding/binary"
)

// Size in bytes of uint64
// https://golang.org/ref/spec#Size_and_alignment_guarantees
const Uint64Size = 8

func pack(m string, s uint64, buf []byte) {
	copy(buf[0:], m)
	copy(buf[len(m):], []byte{
		byte(s),
		byte(s >> 8),
		byte(s >> 16),
		byte(s >> 24),
		byte(s >> 32),
		byte(s >> 40),
		byte(s >> 48),
		byte(s >> 56),
	})
}

func unpack(in []byte, m *string, s *uint64) error {
	*m = string(in[:len(in)-8])
	*s = binary.LittleEndian.Uint64(in[len(in)-8:])

	// no errors for now
	return nil
}
