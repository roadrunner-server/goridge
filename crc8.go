package goridge

import "sync"

// CRC-8 Lookup table
var lookupTable = [256]byte{}

// generator is the constant used to generate lookup table for crc-8
const generator = byte(0x48)

// lookup table should be initialized only once
var once = &sync.Once{}

// Lookup table
func initLookupTable() {
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

func crc8slow(data *[]byte) {
	_ = (*data)[7]
	var crc = byte(0)
	for i := 0; i < len(*data); i++ {
		crc ^= (*data)[i]
		for i := 0; i < 8; i++ {
			if crc&0x80 != 0 {
				crc = (crc << 1) ^ generator
			} else {
				crc <<= 1
			}
		}
	}
	(*data)[6] = crc
	//return crc
}

func crc8slowCheck(data []byte) bool {
	_ = data[7]
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
	return crc == data[6]
}
