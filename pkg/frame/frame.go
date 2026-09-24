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
	// pb is the pooled buffer that backs payload. It is nil when the payload is empty
	// or aliases memory the frame did not take from the pool, and it is the only thing
	// Reset returns to the pool.
	pb *[]byte
}

// ReadHeader reads only the header (first 12 bytes) from data, without payload.
// The frame aliases data: WriteOptions, AppendOptions and Reset write into it.
func ReadHeader(data []byte) *Frame { // inlined, cost 14
	_ = data[11]
	return &Frame{
		header:  data[:12],
		payload: nil,
	}
}

// ReadFrame produces a Frame from raw bytes.
// The first 12 bytes (or more if options are present) form the header; the rest is the payload.
// Without options, ReadFrame clears bytes 10 and 11 of data.
// The frame aliases data: WritePayload, WriteOptions, AppendOptions and Reset write into it,
// and a header extension overwrites the start of the payload.
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

// NewFrame initializes a new frame with a 12-byte header and no payload.
// The header has capacity for the maximum of 10 options, so WriteOptions does not allocate.
// The payload is borrowed from the buffer pool on the first AllocPayload or WritePayload;
// Reset returns it when it is larger than the smallest pool tier.
func NewFrame() *Frame {
	f := &Frame{
		header: make([]byte, 12, 12+OptionsMaxSize),
	}
	// set default header len (2)
	f.defaultHL(f.header)
	return f
}

// From wraps the given header and payload slices as a Frame.
// The frame takes ownership of both slices: WritePayload, WriteOptions, AppendOptions and Reset write into them.
func From(header []byte, payload []byte) *Frame {
	return &Frame{
		payload: payload,
		header:  header,
	}
}

// ReadVersion returns the protocol version from the upper 4 bits of byte 0.
// 1111_0000 -> 0000_1111 (15)
func (*Frame) ReadVersion(header []byte) byte {
	_ = header[0]
	return header[0] >> 4
}

// WriteVersion writes the protocol version (0-15) into the upper 4 bits of byte 0.
// The version is shifted left by 4 and OR'd with the existing byte to preserve the header length.
func (*Frame) WriteVersion(header []byte, version byte) {
	_ = header[0]
	if version > 15 {
		panic("version is only 4 bits")
	}
	header[0] = version<<4 | header[0]
}

// ReadHL reads the header length from the lower 4 bits of byte 0.
// The upper 4 bits (version) are masked off with bitwise AND 0x0F.
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

// ReadFlags returns the flags byte (byte 1) of the frame header.
func (f *Frame) ReadFlags() byte {
	return f.header[1]
}

// WriteFlags sets one or more flags on byte 1 of the header using bitwise OR.
func (*Frame) WriteFlags(header []byte, flags ...byte) {
	_ = header[1]
	for i := range flags {
		header[1] |= flags[i]
	}
}

// SetStreamFlag sets the STREAM bit on byte 10 of the header.
func (*Frame) SetStreamFlag(header []byte) {
	_ = header[11]
	header[10] |= STREAM
}

// IsStream reports whether the STREAM bit is set on byte 10 of the header.
func (*Frame) IsStream(header []byte) bool {
	_ = header[11]
	return header[10]&STREAM != 0
}

// IsPing reports whether the PING bit is set on byte 10 of the header.
func (*Frame) IsPing(header []byte) bool {
	_ = header[11]
	return header[10]&PING != 0
}

// SetPingBit sets the PING bit on byte 10 of the header.
func (*Frame) SetPingBit(header []byte) {
	_ = header[11]
	header[10] |= PING
}

// IsPong reports whether the PONG bit is set on byte 10 of the header.
func (*Frame) IsPong(header []byte) bool {
	_ = header[11]
	return header[10]&PONG != 0
}

// SetPongBit sets the PONG bit on byte 10 of the header.
func (*Frame) SetPongBit(header []byte) {
	_ = header[11]
	header[10] |= PONG
}

// SetStopBit sets the STOP bit on byte 10 of the header.
func (*Frame) SetStopBit(header []byte) {
	_ = header[11]
	header[10] |= STOP
}

// IsStop reports whether the STOP bit is set on byte 10 of the header.
func (*Frame) IsStop(header []byte) bool {
	_ = header[11]
	return header[10]&STOP != 0
}

// WriteOptions writes uint32 option values into the header, extending it by 4 bytes per option.
// At most 10 options (40 bytes) are allowed. The header is extended in place when its capacity is sufficient,
// otherwise it is reallocated, so the header pointer is required.
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

	newSl := growHeader(*header, len(*header)+len(options)*WORD)

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

// ReadOptions reads uint32 option values from the extended header.
// Returns nil if no options are present (HL <= 3). Option count is derived from HL - 3.
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

// ReadPayloadLen reads the payload length from bytes 2-5 of the header in little-endian format.
func (*Frame) ReadPayloadLen(header []byte) uint32 {
	// 2,3,4,5
	_ = header[5]
	return uint32(header[2]) | uint32(header[3])<<8 | uint32(header[4])<<16 | uint32(header[5])<<24
}

// WritePayloadLen writes the payload length into bytes 2-5 of the header in little-endian format.
func (*Frame) WritePayloadLen(header []byte, payloadLen uint32) {
	_ = header[5]
	header[2] = byte(payloadLen)
	header[3] = byte(payloadLen >> 8)
	header[4] = byte(payloadLen >> 16)
	header[5] = byte(payloadLen >> 24)
}

// WriteCRC calculates the CRC32 checksum of bytes 0-5 and writes it into bytes 6-9 of the header.
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

// AppendOptions appends raw option bytes to the header.
// The header is extended in place when its capacity is sufficient, otherwise it is reallocated.
func (*Frame) AppendOptions(header *[]byte, options []byte) {
	newSl := growHeader(*header, len(*header)+len(options))
	// j = 12 - first options byte
	for i, j := 0, 12; i < len(options); i, j = i+1, j+1 {
		newSl[j] = options[i]
	}

	// replace value
	*header = newSl
}

// VerifyCRC verifies the CRC32 checksum stored in bytes 6-9 against the computed checksum of bytes 0-5.
// Returns false if the checksums do not match, indicating a corrupted frame.
func (*Frame) VerifyCRC(header []byte) bool {
	_ = header[9]
	return crc32.ChecksumIEEE(header[:6]) == uint32(header[6])|uint32(header[7])<<8|uint32(header[8])<<16|uint32(header[9])<<24
}

// Bytes returns the full frame as a single byte slice (header + payload).
func (f *Frame) Bytes() []byte {
	buf := make([]byte, 0, len(f.header)+len(f.payload))
	buf = append(buf, f.header...)
	buf = append(buf, f.payload...)
	return buf
}

// Header returns the frame header byte slice.
func (f *Frame) Header() []byte {
	return f.header
}

// HeaderPtr returns a pointer to the frame header slice.
func (f *Frame) HeaderPtr() *[]byte {
	return &f.header
}

// Payload returns the frame payload without the header.
// The slice is valid until the next AllocPayload, WritePayload or Reset on the frame.
func (f *Frame) Payload() []byte {
	return f.payload
}

// AllocPayload sets the payload length to n and returns it for the caller to fill.
// The contents are undefined. The pooled buffer is reused when it fits, a payload that
// aliases caller memory is written in place when it fits, and otherwise a buffer of the
// right tier is taken from the pool. n == 0 never takes a buffer. A pooled payload starts
// Headroom bytes into its buffer, see Wire.
func (f *Frame) AllocPayload(n int) []byte {
	if f.pb != nil && cap(*f.pb)-Headroom >= n {
		f.payload = (*f.pb)[Headroom : Headroom+n]
		return f.payload
	}

	return f.allocPayloadSlow(n)
}

// allocPayloadSlow handles the cases AllocPayload keeps out of its reuse check.
func (f *Frame) allocPayloadSlow(n int) []byte {
	switch {
	case n == 0:
		if f.payload != nil {
			f.payload = f.payload[:0]
		}
	case f.pb == nil && cap(f.payload) >= n:
		f.payload = f.payload[:n]
	default:
		if f.pb != nil {
			putBuf(f.pb)
		}
		f.pb = getBuf(n + Headroom)
		f.payload = (*f.pb)[Headroom : Headroom+n]
	}

	return f.payload
}

// Headroom is the space kept in front of a pooled payload for the header, so that Wire can
// hand the whole frame to a single Write without copying the payload.
const Headroom = 12 + OptionsMaxSize

// Wire writes the header into the headroom in front of a pooled payload and returns the
// contiguous header-plus-payload slice, valid until the next write or Reset. ok is false when
// the frame has no pooled payload, an empty payload, or a header longer than Headroom; the
// caller then sends the header and the payload separately.
func (f *Frame) Wire() (wire []byte, ok bool) {
	if f.pb == nil || len(f.payload) == 0 || len(f.header) > Headroom {
		return nil, false
	}

	start := Headroom - len(f.header)
	copy((*f.pb)[start:Headroom], f.header)

	return (*f.pb)[start : Headroom+len(f.payload)], true
}

// WritePayload copies data into the frame's payload. See AllocPayload for where the buffer comes from.
func (f *Frame) WritePayload(data []byte) {
	copy(f.AllocPayload(len(data)), data)
}

// Reset clears the frame, restoring it to its initial state with a 12-byte header and an empty payload.
// The header keeps its capacity. A payload buffer of the smallest pool tier stays on the frame, so
// small frames do not pay a pool round trip per cycle; a larger one goes back to the pool. Either
// way slices returned earlier by Payload are invalid after Reset. A payload that aliases caller
// memory is only truncated. The header capacity must be at least 12 bytes.
func (f *Frame) Reset() {
	f.header = f.header[:12]
	clear(f.header)
	f.defaultHL(f.header)

	if f.pb != nil && cap(*f.pb) > tier4K {
		putBuf(f.pb)
		f.pb = nil
		f.payload = nil

		return
	}

	f.payload = f.payload[:0]
}

// -------- PRIVATE

// growHeader returns header extended to n bytes. The bytes past the old length are zero.
// The backing array is reused when its capacity is sufficient.
func growHeader(header []byte, n int) []byte {
	if cap(header) >= n {
		clear(header[len(header):n])
		return header[:n]
	}
	newSl := make([]byte, n)
	copy(newSl, header)
	return newSl
}

func (f *Frame) defaultHL(header []byte) {
	f.writeHl(header, 3)
}

// Writing HL is very simple. Since we are using lower 4 bits
// we can easily apply bitwise OR and set lower 4 bits to needed hl value
func (*Frame) writeHl(header []byte, hl byte) {
	header[0] |= hl
}
