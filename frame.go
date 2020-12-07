package goridge

const (
	Version        = 0
	PayloadControl = 1
	PayloadError   = 2
)

// Frame defines new user level package format.
type Frame struct {
	// Payload, max length 4.2GB.
	payload []byte

	// Header
	header []byte
}

func ReadFrame(data []byte) *Frame {
	if lookupTable[1] == 0 {
		panic("should initialize lookup table")
	}
	return &Frame{
		header:  data[:8],
		payload: data[8:],
	}
}

func NewFrame() *Frame {
	if lookupTable[1] == 0 {
		panic("should initialize lookup table")
	}
	return &Frame{
		header:  make([]byte, 8),
		payload: make([]byte, 0, 100),
	}
}

// To read version, we should return our 4 upper bits to their original place
// 1111_0000 -> 0000_1111 (15)
func (f *Frame) ReadVersion() uint8 {
	_ = f.header[0]
	return f.header[0] >> 4
}

// To write version, we should made the following:
// 1. For example we have version 15 it's 0000_1111 (1 byte)
// 2. We should shift 4 lower bits to upper and write that to the 0th byte
// 3. The 0th byte should become 1111_0000, but it's not 240, it's only 15, because version only 4 bits len
func (f *Frame) WriteVersion(version uint8) {
	_ = f.header[0]
	if version > 15 {
		panic("version is only 4 bits")
	}
	f.header[0] = version << 4
}

// The lower 4 bits of the 0th octet occupies our header len data.
// We should erase upper 4 bits, which contain information about Version
// To erase, we applying bitwise AND to the upper 4 bits and returning result
func (f *Frame) ReadHL() uint8 {
	_ = f.header[0]
	// 0101_1111         0000_1111
	return f.header[0] & uint8(0x0F)
}

// Writing HL is very simple. Since we are using lower 4 bits
// we can easily apply bitwise OR and set lower 4 bits to needed hl value
func (f *Frame) WriteHL(hl uint8) {
	_ = f.header[0]
	if hl > 15 {
		panic("header length is only 4 bits")
	}
	//
	f.header[0] = f.header[0] | hl
}

// Flags is full 1st byte
func (f *Frame) ReadFlags() uint8 {
	_ = f.header[1]
	return f.header[1]
}

func (f *Frame) WriteFlags(flags uint8) {
	_ = f.header[1]
	f.header[1] = f.header[1] | flags
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


func (f *Frame) WritePayload(data []byte) {
	f.payload = make([]byte, len(data))
	f.payload = data
}
