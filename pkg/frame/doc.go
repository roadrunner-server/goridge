// Package frame implements a binary frame protocol for inter-process communication.
//
// Each frame consists of a header (12 bytes minimum) followed by a variable-length payload.
// The header is laid out as follows:
//
//	Byte 0:       version (upper 4 bits) | header length in 32-bit words (lower 4 bits)
//	Byte 1:       flags — codec selection (CodecRaw, CodecJSON, CodecMsgpack, CodecGob, CodecProto)
//	              and control flags (CONTROL, ERROR)
//	Bytes 2-5:    payload length (little-endian uint32, max ~4 GB)
//	Bytes 6-9:    CRC32 (IEEE) of bytes 0-5
//	Bytes 10-11:  stream control bits (STREAM, STOP, PING, PONG)
//
// When header length exceeds 3 words (12 bytes), the additional words carry
// up to 10 option values (uint32 each, 40 bytes maximum) appended after the
// base header.
package frame
