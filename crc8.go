package goridge

import "sync"

// CRC-8 Lookup table
var lookupTable = [256]byte{}

var once = &sync.Once{}

func createLookupTable() {
	once.Do(func() {
		for i := 0; i < 256; i++ {
			currByte := byte(i)

			for j := 0; j < 8; j++ {
				if (currByte & 0x80) != 0 {
					currByte = currByte << 1
					currByte ^= generator
				} else {
					currByte = currByte << 1
				}
			}

			lookupTable[i] = currByte
		}
	})
}

const generator = byte(0x1F)

func crc8slow(data []byte) byte {
	var crc = byte(0)
	for i := 0; i < len(data); i++ {
		crc ^= data[i]
		for i := 0; i < 8; i++ {
			if crc&0x80 != 0 {
				crc = (crc << 1) ^ generator
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}
