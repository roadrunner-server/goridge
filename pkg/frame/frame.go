package frame

import "hash/crc32"

// OptionsMaxSize represents header's options maximum size
const OptionsMaxSize = 40

// WORD represents 32bit word
const WORD = 4

// Frame defines new user level package format.
type Frame struct {
	// Payload, max length 4.2GB.
	payload []byte

	// Header
	header []byte
}

// ReadHeader reads only header, without payload
func ReadHeader(data []byte) *Frame {
	_ = data[11]
	return &Frame{
		header:  data[:12],
		payload: nil,
	}
}

// ReadFrame produces Frame from the RAW bytes
// first 12 bytes will be a header
// the rest - payload
func ReadFrame(data []byte) *Frame {
	_ = data[0]
	opt := data[0] & 0x0F
	// if more than 3, that we have options
	if opt > 3 {
		return &Frame{
			header:  data[:opt*WORD],
			payload: data[opt*WORD:],
		}
	}

	f := &Frame{
		header:  data[:12],
		payload: data[12:],
	}
	f.header[10] = 0
	f.header[11] = 0

	return f
}

// NewFrame initializes new frame with 12-byte header and 100-byte reserved space for the payload
func NewFrame() *Frame {
	f := &Frame{
		header:  make([]byte, 12),
		payload: make([]byte, 0, 100),
	}
	// set default header len (2)
	f.defaultHL()
	return f
}

// From .. MergeHeader merge header from other frame with original payload
func From(header []byte, payload []byte) *Frame {
	return &Frame{
		payload: payload,
		header:  header,
	}
}

// ReadVersion .. To read version, we should return our 4 upper bits to their original place
// 1111_0000 -> 0000_1111 (15)
//go:inline
func (f *Frame) ReadVersion() byte {
	_ = f.header[0]
	return f.header[0] >> 4
}

// WriteVersion ..
// To write version, we should made the following:
// 1. For example we have version 15 it's 0000_1111 (1 byte)
// 2. We should shift 4 lower bits to upper and write that to the 0th byte
// 3. The 0th byte should become 1111_0000, but it's not 240, it's only 15, because version only 4 bits len
//go:inline
func (f *Frame) WriteVersion(version Version) {
	_ = f.header[0]
	if version > 15 {
		panic("version is only 4 bits")
	}
	f.header[0] = byte(version)<<4 | f.header[0]
}

// ReadHL ..
// The lower 4 bits of the 0th octet occupies our header len data.
// We should erase upper 4 bits, which contain information about Version
// To erase, we applying bitwise AND to the upper 4 bits and returning result
//go:inline
func (f *Frame) ReadHL() byte {
	_ = f.header[0]
	// 0101_1111         0000_1111
	return f.header[0] & 0x0F
}

// Writing HL is very simple. Since we are using lower 4 bits
// we can easily apply bitwise OR and set lower 4 bits to needed hl value
//go:inline
func (f *Frame) writeHl(hl byte) {
	_ = f.header[0]
	f.header[0] |= hl
}

//go:inline
func (f *Frame) incrementHL() {
	_ = f.header[0]
	hl := f.ReadHL()
	if hl > 15 {
		panic("header len should be less than 15")
	}
	f.header[0] = f.header[0] | hl + 1
}

//go:inline
func (f *Frame) defaultHL() {
	_ = f.header[0]
	f.writeHl(3)
}

// ReadFlags ..
// Flags is full 1st byte
//go:inline
func (f *Frame) ReadFlags() byte {
	_ = f.header[1]
	return f.header[1]
}

func (f *Frame) WriteFlags(flags ...byte) {
	_ = f.header[1]
	for i := 0; i < len(flags); i++ {
		f.header[1] |= flags[i]
	}
}

// WriteOptions ..
// Options slice len should not be more than 10 (40 bytes)
func (f *Frame) WriteOptions(options ...uint32) {
	if options == nil {
		return
	}
	if len(options) > 10 {
		panic("header options limited by 40 bytes")
	}

	hl := f.ReadHL()
	// check before writing. we can't handle more than 15*4 bytes of HL (3 for header and 12 for options)
	if hl == 15 {
		panic("header len could not be more than 15")
	}

	tmp := make([]byte, 0, OptionsMaxSize)
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
//go:inline
func (f *Frame) AppendOptions(opts []byte) {
	f.header = append(f.header, opts...)
}

// last byte after main header and first options byte
const lb = 12

// ReadOptions ...
// f.readHL() - 2 needed to know actual options size
// we know, that 2 WORDS is minimal header len
// extra WORDS will add extra 32bits to the options (4 bytes)
func (f *Frame) ReadOptions() []uint32 {
	// we can read options, if there are no options
	if f.ReadHL() <= 3 {
		return nil
	}
	// Get the options len
	optionLen := f.ReadHL() - 3 // 3 is the default
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

// ReadPayloadLen ..
// LE format used to write Payload
// Using 4 bytes (2,3,4,5 bytes in the header)
//go:inline
func (f *Frame) ReadPayloadLen() uint32 {
	// 2,3,4,5
	_ = f.header[5]
	return uint32(f.header[2]) | uint32(f.header[3])<<8 | uint32(f.header[4])<<16 | uint32(f.header[5])<<24
}

// WritePayloadLen ..
// LE format used to write Payload
// Using 4 bytes (2,3,4,5 bytes in the header)
//go:inline
func (f *Frame) WritePayloadLen(len uint32) {
	_ = f.header[5]
	f.header[2] = byte(len)
	f.header[3] = byte(len >> 8)
	f.header[4] = byte(len >> 16)
	f.header[5] = byte(len >> 24)
}

// WriteCRC ..
// Calculating CRC and writing it to the 6th byte (7th reserved)
func (f *Frame) WriteCRC() {
	// 6 7 8 9 bytes
	// 10, 11 reserved
	_ = f.header[9]

	crc := crc32.ChecksumIEEE(f.header[:6])
	f.header[6] = byte(crc)
	f.header[7] = byte(crc >> 8)
	f.header[8] = byte(crc >> 16)
	f.header[9] = byte(crc >> 24)
}

// VerifyCRC ..
// Reading info from 6th byte and verifying it with calculated in-place. Should be equal.
// If not - drop the frame as incorrect.
func (f *Frame) VerifyCRC() bool {
	_ = f.header[9]

	return crc32.ChecksumIEEE(f.header[:6]) == uint32(f.header[6])|uint32(f.header[7])<<8|uint32(f.header[8])<<16|uint32(f.header[9])<<24
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

// WritePayload .. writes payload
func (f *Frame) WritePayload(data []byte) {
	f.payload = make([]byte, len(data))
	copy(f.payload, data)
}
