package goridge

import (
	"unsafe"
)

const FRAME_OPTIONS_MAX_SIZE = 40 //nolint:golint
const WORD = 4                    //nolint:golint

// Frame defines new user level package format.
type Frame struct {
	// Payload, max length 4.2GB.
	payload []byte

	// Header
	header []byte
}

// ReadHeader reads only header, without payload
func ReadHeader(data []byte) *Frame {
	if lookupTable[1] == 0 {
		panic("should initialize lookup table")
	}
	_ = data[7]
	return &Frame{
		header:  data[:8],
		payload: nil,
	}
}

func ReadFrame(data []byte) *Frame {
	if lookupTable[1] == 0 {
		panic("should initialize lookup table")
	}
	_ = data[0]
	opt := data[0] & 0x0F
	// if more than 2, that we have options
	if opt > 2 {
		return &Frame{
			header:  data[:opt*WORD],
			payload: data[opt*WORD:],
		}
	}

	// no options
	return &Frame{
		header:  data[:8],
		payload: data[8:],
	}
}

func NewFrame() *Frame {
	initLookupTable()
	if lookupTable[1] == 0 {
		panic("should initialize lookup table")
	}
	f := &Frame{
		header:  make([]byte, 8),
		payload: make([]byte, 0, 100),
	}
	// set default header len (2)
	f.defaultHL()
	return f
}

// MergeHeader merge header from other frame with original payload
func (f *Frame) MergeHeader(frame *Frame) {
	f.header = frame.header
}

// To read version, we should return our 4 upper bits to their original place
// 1111_0000 -> 0000_1111 (15)
func (f *Frame) ReadVersion() byte {
	_ = f.header[0]
	return f.header[0] >> 4
}

// To write version, we should made the following:
// 1. For example we have version 15 it's 0000_1111 (1 byte)
// 2. We should shift 4 lower bits to upper and write that to the 0th byte
// 3. The 0th byte should become 1111_0000, but it's not 240, it's only 15, because version only 4 bits len
func (f *Frame) WriteVersion(version Version) {
	_ = f.header[0]
	if version > 15 {
		panic("version is only 4 bits")
	}
	f.header[0] = byte(version)<<4 | f.header[0]
}

// The lower 4 bits of the 0th octet occupies our header len data.
// We should erase upper 4 bits, which contain information about Version
// To erase, we applying bitwise AND to the upper 4 bits and returning result
func (f *Frame) readHL() byte {
	_ = f.header[0]
	// 0101_1111         0000_1111
	return f.header[0] & 0x0F
}

// Writing HL is very simple. Since we are using lower 4 bits
// we can easily apply bitwise OR and set lower 4 bits to needed hl value
func (f *Frame) writeHl(hl byte) {
	_ = f.header[0]
	f.header[0] = f.header[0] | hl
}

func (f *Frame) incrementHL() {
	_ = f.header[0]
	hl := f.readHL()
	if hl > 15 {
		panic("header len should be less than 15")
	}
	f.header[0] = f.header[0] | hl + 1
}
func (f *Frame) defaultHL() {
	_ = f.header[0]
	f.writeHl(2)
}

// Flags is full 1st byte
func (f *Frame) ReadFlags() byte {
	_ = f.header[1]
	return f.header[1]
}

func (f *Frame) WriteFlags(flags ...FrameFlag) {
	_ = f.header[1]
	for i := 0; i < len(flags); i++ {
		f.header[1] = f.header[1] | byte(flags[i])
	}
}

// Options slice len should not be more than 10 (40 bytes)
func (f *Frame) WriteOptions(options ...uint32) {
	if options == nil {
		panic("you should write at least one option (uint32)")
	}
	hl := f.readHL()
	// check before writing. we can't handle more than 15*4 bytes of HL (2 for header and 12 for options)
	if hl == 15 {
		panic("header len could not be more than 15")
	}
	if len(options) > 10 {
		panic("header options limited by 40 bytes")
	}

	tmp := make([]byte, 0, FRAME_OPTIONS_MAX_SIZE)
	for i := 0; i < len(options); i++ {
		tmp = append(tmp, byte(options[i]))
		tmp = append(tmp, byte(options[i]>>8))
		tmp = append(tmp, byte(options[i]>>16))
		tmp = append(tmp, byte(options[i]>>24))
		f.incrementHL() // increment header len by 32 bit
	}

	f.header = append(f.header, tmp...)
}

// AppendOptions appends options to the header
func (f *Frame) AppendOptions(opts []byte) {
	f.header = append(f.header, opts...)
}

// last byte after main header and first options byte
const lb = 8

// f.readHL() - 2 needed to know actual options size
// we know, that 2 WORDS is minimal header len
// extra WORDS will add extra 32bits to the options (4 bytes)
func (f *Frame) ReadOptions() []uint32 {
	// we can read options, if there are no options
	if f.readHL() <= 2 {
		return nil
	}
	// Get the options len
	optionLen := f.readHL() - 2 // 2 is the default
	// slice in place
	options := make([]uint32, 0, optionLen)

	// Options starting from 8-th byte
	// we should scan with 4 byte window (32bit, WORD)
	for i := byte(0); i < optionLen*WORD; i += WORD {
		// for example
		// 8  12  16
		// 9  13  17
		// 10 14  18
		// 11 15  19
		// For this data, HL will be 3, optionLen will be 12 (3*4) bytes
		options = append(options, uint32(f.header[lb+i])|uint32(f.header[lb+i+1])<<8|uint32(f.header[lb+i+2])<<16|uint32(f.header[lb+i+3])<<24)
	}
	return options
}

// LE format used to write Payload
// Using 4 bytes (2,3,4,5 bytes in the header)
func (f *Frame) ReadPayloadLen() uint32 {
	// 2,3,4,5
	_ = f.header[5]
	return uint32(f.header[2]) | uint32(f.header[3])<<8 | uint32(f.header[4])<<16 | uint32(f.header[5])<<24
}

// LE format used to write Payload
// Using 4 bytes (2,3,4,5 bytes in the header)
func (f *Frame) WritePayloadLen(len uint32) {
	_ = f.header[5]
	f.header[2] = byte(len)
	f.header[3] = byte(len >> 8)
	f.header[4] = byte(len >> 16)
	f.header[5] = byte(len >> 24)
}

// Calculating CRC and writing it to the 6th byte (7th reserved)
func (f *Frame) WriteCRC() {
	_ = f.header[7]
	crc := byte(0)
	hl := f.readHL()
	// write CRC with options
	if f.readHL() > 2 {
		for i := byte(0); i < hl*WORD; i++ {
			data := f.header[i] ^ crc
			crc = lookupTable[data]
		}
		f.header[6] = crc
		return
	}

	for i := 0; i < 6; i++ {
		data := f.header[i] ^ crc
		crc = lookupTable[data]
	}

	f.header[6] = crc
}

// Reading info from 6th byte and verifying it with calculated in-place. Should be equal.
// If not - drop the frame as incorrect.
func (f *Frame) VerifyCRC() bool {
	_ = f.header[7]
	crc := byte(0)
	hl := f.readHL()

	if hl > 2 {
		for i := byte(0); i < hl*WORD; i++ {
			// to verify, we are skipping the CRC field itself
			if i == 6 {
				data := 0 ^ crc
				crc = lookupTable[data]
				continue
			}
			data := f.header[i] ^ crc
			crc = lookupTable[data]
		}
		return crc == f.header[6]
	}

	for i := 0; i < 6; i++ {
		data := f.header[i] ^ crc
		crc = lookupTable[data]
	}
	return crc == f.header[6]
}

// Bytes returns header with payload
func (f *Frame) Bytes() []byte {
	buf := make([]byte, 0, len(f.header)+len(f.payload))
	buf = append(buf, f.header...)
	buf = append(buf, f.payload...)
	return buf
}

// Header returns frame header
func (f *Frame) Header() []byte {
	return f.header
}

// Payload returns frame payload without header
func (f *Frame) Payload() []byte {
	// start from the 1st (staring from 0) byte
	return f.payload
}

//
func (f *Frame) WritePayload(data []byte) {
	f.payload = make([]byte, len(data))
	copy(f.payload, data)
}

//go:nosplit
//go:nocheckptr
func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0) //nolint:staticcheck
}

// After reset you should write all data from the start
func (f *Frame) Reset() {
	f.header = make([]byte, 0, 8)
	f.payload = make([]byte, 0, 100)
}
