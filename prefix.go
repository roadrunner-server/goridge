// Author Wolfy-J, 2017. License MIT

package goridge

import (
	"encoding/binary"
)

// Prefix is always 9 bytes long and contain meta flags and length of next data package.
type Prefix []byte

func (p Prefix) Flags() byte {
	return p[0]
}

func (p Prefix) SetFlags(flags byte) {
	p[0] = p[0] | flags
}

func (p Prefix) Size() uint64 {
	return binary.LittleEndian.Uint64(p[1:])
}

func (p Prefix) HasBody() bool {
	return p.Flags()&NoBody == 0 && p.Size() != 0
}

func (p Prefix) SetSize(size uint64) {
	binary.LittleEndian.PutUint64(p[1:], size)
}
