package goridge

import (
	"encoding/binary"
	"fmt"
)

const (
	// PayloadEmpty must be set when no data to be sent.
	PayloadEmpty byte = 2

	// PayloadRaw must be set when data binary data.
	PayloadRaw byte = 4

	// PayloadError must be set when data is error string or structure.
	PayloadError byte = 8

	// PayloadControl defines that associated data must be treated as control data.
	PayloadControl byte = 16
)

// Prefix is always 17 bytes long and contain meta flags and length of next data package. Receive prefix by converting it
// into the slice. Prefix duplicates size using reverse bytes order to detect possible transmission errors.
type Prefix [17]byte

// NewPrefix creates new empty prefix with no flags and size.
func NewPrefix() Prefix {
	return Prefix([17]byte{})
}

// String represents prefix as string
func (p Prefix) String() string {
	return fmt.Sprintf("[%08b: %v]", p.Flags(), p.Size())
}

// Flags describe transmission behaviour and data data type.
func (p Prefix) Flags() byte {
	return p[0]
}

// HasFlag returns true if prefix has given flag.
func (p Prefix) HasFlag(flag byte) bool {
	return p[0]&flag == flag
}

// Valid returns true if prefix is valid.
func (p Prefix) Valid() bool {
	return binary.LittleEndian.Uint64(p[1:]) == binary.BigEndian.Uint64(p[9:])
}

// Size returns following data size in bytes.
func (p Prefix) Size() uint64 {
	if p.HasFlag(PayloadEmpty) {
		return 0
	}

	return binary.LittleEndian.Uint64(p[1:])
}

// HasPayload returns true if data is not empty.
func (p Prefix) HasPayload() bool {
	return p.Size() != 0
}

// WithFlag unites given value with flag byte and returns new instance of prefix.
func (p Prefix) WithFlag(flag byte) Prefix {
	p[0] = p[0] | flag
	return p
}

// WithFlags overwrites all flags and returns new instance of prefix.
func (p Prefix) WithFlags(flags byte) Prefix {
	p[0] = flags
	return p
}

// WithSize returns new prefix with given size.
func (p Prefix) WithSize(size uint64) Prefix {
	binary.LittleEndian.PutUint64(p[1:], size)
	binary.BigEndian.PutUint64(p[9:], size)
	return p
}
