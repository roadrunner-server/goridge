package frame

// BYTE flags, it means, that we can set multiply flags from this group using bitwise OR
// For example CONTEXT_SEPARATOR | CodecRaw
const (
	CONTROL      byte = 0x01
	CodecRaw     byte = 0x04
	CodecJSON    byte = 0x08
	CodecMsgpack byte = 0x10
	CodecGob     byte = 0x20
	ERROR        byte = 0x40
	CodecProto   byte = 0x80

	// Version1 byte
	Version1 byte = 0x01

	/*
		10th byte
	*/

	// STREAM bit
	STREAM byte = 0x01
	// STOP command
	STOP byte = 0x02
)
