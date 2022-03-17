package frame

import (
	"hash/crc32"
)

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
func ReadHeader(data []byte) *Frame { // inlined, cost 14
	_ = data[11]
	return &Frame{
		header:  data[:12],
		payload: nil,
	}
}

// ReadFrame produces Frame from the RAW bytes
// first 12 bytes will be a header
// the rest - payload
func ReadFrame(data []byte) *Frame { // inlined, cost 60
	_ = data[11]
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

	f.header[10], f.header[11] = 0, 0

	return f
}

// NewFrame initializes new frame with 12-byte header and 100-byte reserved space for the payload
func NewFrame() *Frame {
	f := &Frame{
		header:  make([]byte, 12),
		payload: make([]byte, 0, 100),
	}
	// set default header len (2)
	f.defaultHL(f.header)
	return f
}

// From will represent header and payload as a Frame
func From(header []byte, payload []byte) *Frame {
	return &Frame{
		payload: payload,
		header:  header,
	}
}

// ReadVersion ... To read version, we should return our 4 upper bits to their original place
// 1111_0000 -> 0000_1111 (15)
func (*Frame) ReadVersion(header []byte) byte {
	_ = header[0]
	return header[0] >> 4
}

// WriteVersion
// To write version, we should do the following:
// 1. For example, we have version 15 it's 0000_1111 (1 byte)
// 2. We should shift 4 lower bits to upper and write that to the 0th byte
// 3. The 0th byte should become 1111_0000, but it's not 240, it's only 15, because version only 4 bits len
func (*Frame) WriteVersion(header []byte, version byte) {
	_ = header[0]
	if version > 15 {
		panic("version is only 4 bits")
	}
	header[0] = version<<4 | header[0]
}

// ReadHL
// The lower 4 bits of the 0th octet occupies our header len data.
// We should erase upper 4 bits, which contain information about Version
// To erase, we are applying bitwise AND to the upper 4 bits and returning result
func (*Frame) ReadHL(header []byte) byte {
	// 0101_1111         0000_1111
	// 0x0F - 15
	return header[0] & 0x0F
}

func (f *Frame) incrementHL(header []byte) {
	hl := f.ReadHL(header)
	if hl == 15 {
		panic("header len should be less than 15 to increment")
	}
	header[0] = header[0] | hl + 1
}

// ReadFlags
// Flags is full 1st byte
func (f *Frame) ReadFlags() byte {
	return f.header[1]
}

func (*Frame) WriteFlags(header []byte, flags ...byte) {
	_ = header[1]
	for i := 0; i < len(flags); i++ {
		header[1] |= flags[i]
	}
}

func (*Frame) SetStreamFlag(header []byte) {
	_ = header[11]
	header[10] |= STREAM
}

func (*Frame) IsStream(header []byte) bool {
	_ = header[11]
	return header[10]&STREAM == 1
}

func (*Frame) SetStopBit(header []byte) {
	header[10] |= STOP
}

func (*Frame) IsStop(header []byte) bool {
	return header[10]&STOP != 0
}

// WriteOptions
// Options slice len should not be more than 10 (40 bytes)
// we need a pointer to the header because we are reallocating the slice
func (f *Frame) WriteOptions(header *[]byte, options ...uint32) {
	if options == nil {
		return
	}

	if header == nil {
		panic("header should not be nil")
	}

	if len(options) > 10 {
		panic("header options limited by 40 bytes")
	}

	hl := f.ReadHL(*header)
	// check before writing. we can't handle more than 15*4 bytes of HL (3 for header and 12 for options)
	if hl == 15 {
		panic("header len could not be equal to 15 to write options")
	}

	// make a new slice with the exact len (not doubled)
	newSl := make([]byte, (len(options)*WORD)+len(*header))
	// copy old data
	copy(newSl, *header)

	for i, j := 0, 12; i < len(options); i, j = i+1, j+WORD {
		newSl[j] |= byte(options[i])
		newSl[j+1] |= byte(options[i] >> 8)
		newSl[j+2] |= byte(options[i] >> 16)
		newSl[j+3] |= byte(options[i] >> 24)

		f.incrementHL(newSl) // increment header len by 32 bit
	}

	// replace value
	*header = newSl
}

// ReadOptions
// f.readHL() - 2 needed to know actual options size
// we know, that 2 WORDS is minimal header len
// extra WORDS will add extra 32bits to the options (4 bytes)
// cannot inline, cost 117 vs 80
func (f *Frame) ReadOptions(header []byte) []uint32 { //nolint:funlen
	ol := f.ReadHL(header)
	// we can read options, if there are no options
	if ol <= 3 {
		return nil
	}

	// last byte after main header and first options byte
	const lb = 12

	// Get the options len minus the standard options
	optionLen := ol - 3 // 3 is the default
	// check the options len
	if optionLen*WORD > OptionsMaxSize {
		panic("options size is limited by 40 bytes (10 4-bytes words)")
	}

	// slice in place
	options := make([]uint32, optionLen)

	// SAMPLE
	// Options starting from 12-th byte till 52-th byte (40 bytes max)
	// we should scan with 4 byte window (32bit, WORD)
	// 8  12  16
	// 9  13  17
	// 10 14  18
	// 11 15  19

	// loop unwind
	i := byte(0)
	j := 0

	_ = header[lb+i+3]
	_ = options[j]

	// 1
	options[j] |= uint32(header[lb+i])
	options[j] |= uint32(header[lb+i+1]) << 8
	options[j] |= uint32(header[lb+i+2]) << 16
	options[j] |= uint32(header[lb+i+3]) << 24

	i += WORD
	j++

	if i == optionLen*WORD {
		goto done
	}

	_ = header[lb+i+3]
	_ = options[j]
	// 2
	options[j] |= uint32(header[lb+i])
	options[j] |= uint32(header[lb+i+1]) << 8
	options[j] |= uint32(header[lb+i+2]) << 16
	options[j] |= uint32(header[lb+i+3]) << 24

	i += WORD
	j++

	if i == optionLen*WORD {
		goto done
	}

	_ = header[lb+i+3]
	_ = options[j]
	// 3
	options[j] |= uint32(header[lb+i])
	options[j] |= uint32(header[lb+i+1]) << 8
	options[j] |= uint32(header[lb+i+2]) << 16
	options[j] |= uint32(header[lb+i+3]) << 24

	i += WORD
	j++

	if i == optionLen*WORD {
		goto done
	}

	_ = header[lb+i+3]
	_ = options[j]
	// 4
	options[j] |= uint32(header[lb+i])
	options[j] |= uint32(header[lb+i+1]) << 8
	options[j] |= uint32(header[lb+i+2]) << 16
	options[j] |= uint32(header[lb+i+3]) << 24

	i += WORD
	j++

	if i == optionLen*WORD {
		goto done
	}

	_ = header[lb+i+3]
	_ = options[j]
	// 5
	options[j] |= uint32(header[lb+i])
	options[j] |= uint32(header[lb+i+1]) << 8
	options[j] |= uint32(header[lb+i+2]) << 16
	options[j] |= uint32(header[lb+i+3]) << 24

	i += WORD
	j++

	if i == optionLen*WORD {
		goto done
	}

	_ = header[lb+i+3]
	_ = options[j]
	// 6
	options[j] |= uint32(header[lb+i])
	options[j] |= uint32(header[lb+i+1]) << 8
	options[j] |= uint32(header[lb+i+2]) << 16
	options[j] |= uint32(header[lb+i+3]) << 24

	i += WORD
	j++

	if i == optionLen*WORD {
		goto done
	}

	_ = header[lb+i+3]
	_ = options[j]
	// 7
	options[j] |= uint32(header[lb+i])
	options[j] |= uint32(header[lb+i+1]) << 8
	options[j] |= uint32(header[lb+i+2]) << 16
	options[j] |= uint32(header[lb+i+3]) << 24

	i += WORD
	j++

	if i == optionLen*WORD {
		goto done
	}

	_ = header[lb+i+3]
	_ = options[j]
	// 8
	options[j] |= uint32(header[lb+i])
	options[j] |= uint32(header[lb+i+1]) << 8
	options[j] |= uint32(header[lb+i+2]) << 16
	options[j] |= uint32(header[lb+i+3]) << 24

	i += WORD
	j++

	if i == optionLen*WORD {
		goto done
	}

	_ = header[lb+i+3]
	_ = options[j]
	// 9
	options[j] |= uint32(header[lb+i])
	options[j] |= uint32(header[lb+i+1]) << 8
	options[j] |= uint32(header[lb+i+2]) << 16
	options[j] |= uint32(header[lb+i+3]) << 24

	i += WORD
	j++

	if i == optionLen*WORD {
		goto done
	}

	_ = header[lb+i+3]
	_ = options[j]
	// 10 - last possible
	options[j] |= uint32(header[lb+i])
	options[j] |= uint32(header[lb+i+1]) << 8
	options[j] |= uint32(header[lb+i+2]) << 16
	options[j] |= uint32(header[lb+i+3]) << 24

	return options

done:
	return options
}

// ReadPayloadLen
// LE format used to write Payload
// Using 4 bytes (2,3,4,5 bytes in the header)
func (*Frame) ReadPayloadLen(header []byte) uint32 {
	// 2,3,4,5
	_ = header[5]
	return uint32(header[2]) | uint32(header[3])<<8 | uint32(header[4])<<16 | uint32(header[5])<<24
}

// WritePayloadLen
// LE format used to write Payload
// Using 4 bytes (2,3,4,5 bytes in the header)
func (*Frame) WritePayloadLen(header []byte, payloadLen uint32) {
	_ = header[5]
	header[2] = byte(payloadLen)
	header[3] = byte(payloadLen >> 8)
	header[4] = byte(payloadLen >> 16)
	header[5] = byte(payloadLen >> 24)
}

// WriteCRC will calculate and write CRC32 4-bytes it to the 6th byte (7th reserved)
func (*Frame) WriteCRC(header []byte) {
	// 6 7 8 9 10 11 bytes
	_ = header[11]
	// calculate crc
	crc := crc32.ChecksumIEEE(header[:6])
	header[6] = byte(crc)
	header[7] = byte(crc >> 8)
	header[8] = byte(crc >> 16)
	header[9] = byte(crc >> 24)
}

// AppendOptions appends options to the header
func (*Frame) AppendOptions(header *[]byte, options []byte) {
	// make a new slice with the exact len (not doubled)
	newSl := make([]byte, len(options)+len(*header))
	// copy old data
	copy(newSl, *header)
	// j = 12 - first options byte
	for i, j := 0, 12; i < len(options); i, j = i+1, j+1 {
		newSl[j] = options[i]
	}

	// replace value
	*header = newSl
}

// VerifyCRC ...
// Reading info from 6th byte and verifying it with calculated in-place. Should be equal.
// If not - drop the frame as incorrect.
func (*Frame) VerifyCRC(header []byte) bool {
	_ = header[9]
	return crc32.ChecksumIEEE(header[:6]) == uint32(header[6])|uint32(header[7])<<8|uint32(header[8])<<16|uint32(header[9])<<24
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

// HeaderPtr returns frame header pointer
func (f *Frame) HeaderPtr() *[]byte {
	return &f.header
}

// Payload returns frame payload without header
func (f *Frame) Payload() []byte {
	// start from the 1st (staring from 0) byte
	return f.payload
}

// WritePayload writes payload
func (f *Frame) WritePayload(data []byte) {
	f.payload = make([]byte, len(data))
	copy(f.payload, data)
}

// Reset a frame
func (f *Frame) Reset() {
	f.header = make([]byte, 12)
	f.payload = make([]byte, 0, 100)

	f.defaultHL(f.header)
}

// -------- PRIVATE
func (f *Frame) defaultHL(header []byte) {
	f.writeHl(header, 3)
}

// Writing HL is very simple. Since we are using lower 4 bits
// we can easily apply bitwise OR and set lower 4 bits to needed hl value
func (*Frame) writeHl(header []byte, hl byte) {
	header[0] |= hl
}
