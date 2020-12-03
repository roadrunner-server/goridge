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
	return &Frame{
		header:  data[:8],
		payload: data[8:],
	}
}

func NewFrame() *Frame {
	return &Frame{
		header:  make([]byte, 8),
		payload: make([]byte, 0, 100),
	}
}

func (f *Frame) ReadVersion() uint8 {
	_ = f.header[0]
	return f.header[0] >> 4
}

func (f *Frame) WriteVersion(version uint8) {
	_ = f.header[0]
	if version > 15 {
		panic("version is only 4 bits")
	}
	f.header[0] = version << 4
}

// Erasing upper 4 bits
// 0000_1111
func (f *Frame) ReadHL() uint8 {
	_ = f.header[0]
	return f.header[0] & uint8(15)
}

func (f *Frame) WriteHL(hl uint8) {
	_ = f.header[0]
	if hl > 15 {
		panic("header length is only 4 bits")
	}
	f.header[0] = f.header[0] | hl
}

func (f *Frame) ReadFlags() uint8 {
	_ = f.header[1]
	return f.header[1]
}

func (f *Frame) WriteFlags(flags uint8) {
	_ = f.header[1]
	f.header[1] = f.header[1] | flags
}

// LE format
func (f *Frame) ReadPayloadLen() uint32 {
	// 2,3,4,5
	_ = f.header[5]
	return uint32(f.header[2]) | uint32(f.header[3])<<8 | uint32(f.header[4])<<16 | uint32(f.header[5])<<24
}

// LE format
func (f *Frame) WritePayloadLen(len uint32) {
	_ = f.header[5]
	f.header[2] = byte(len)
	f.header[3] = byte(len >> 8)
	f.header[4] = byte(len >> 16)
	f.header[5] = byte(len >> 24)
}

func (f *Frame) WriteCRC(useLookupTable bool) {
	_ = f.header[7]
	crc := byte(0)
	if useLookupTable {
		for i := 0; i < 6; i++ {
			data := f.header[i] ^ crc
			crc = lookupTable[data]
		}

		f.header[6] = crc
		return
	}

	f.header[6] = crc8slow(f.header[:6])
}

func (f *Frame) VerifyCRC(useLookupTable bool) bool {
	_ = f.header[7]
	crc := byte(0)

	if useLookupTable {
		for i := 0; i < 6; i++ {
			data := f.header[i] ^ crc
			crc = lookupTable[data]
		}
		return crc == f.header[6]
	}

	return crc8slow(f.header[:6]) == f.header[6]
}

func (f *Frame) Bytes() []byte {
	buf := make([]byte, 0, len(f.header)+len(f.payload))
	buf = append(buf, f.header...)
	buf = append(buf, f.payload...)
	return buf
}

func (f *Frame) Header() []byte {
	return f.header
}

func (f *Frame) ReadPayload() []byte {
	// start from the 1st (staring from 0) byte
	return f.payload
}

func (f *Frame) WritePayload(data []byte) {
	f.payload = make([]byte, len(data))
	f.payload = data
}
