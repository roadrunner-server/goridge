package goridge

import (
	"bytes"
	"encoding/binary"
)

func pack(m string, s uint64) []byte {
	b := bytes.Buffer{}
	b.WriteString(m)

	b.Write([]byte{
		byte(s),
		byte(s >> 8),
		byte(s >> 16),
		byte(s >> 24),
		byte(s >> 32),
		byte(s >> 40),
		byte(s >> 48),
		byte(s >> 56),
	})

	return b.Bytes()
}

func unpack(in []byte, m *string, s *uint64) error {
	*m = string(in[:len(in)-8])
	*s = binary.LittleEndian.Uint64(in[len(in)-8:])

	// no errors for now
	return nil
}
