package frame

// BYTE flags, it means, that we can set multiply flags from this group using bitwise OR
// For example CONTEXT_SEPARATOR | CODEC_RAW
const (
	CONTROL       byte = 0x01
	CODEC_RAW     byte = 0x04 //nolint:stylecheck,golint
	CODEC_JSON    byte = 0x08 //nolint:stylecheck,golint
	CODEC_MSGPACK byte = 0x10 //nolint:stylecheck,golint
	CODEC_GOB     byte = 0x20 //nolint:stylecheck,golint
	ERROR         byte = 0x40 //nolint:stylecheck,golint
	CODEC_PROTO   byte = 0x80 //nolint:stylecheck,golint
)

// COMPLEX flags can't be used with Byte flags, because it's value more than 128
const (
	RESERVED2 byte = 0x81
	RESERVED3 byte = 0x82
	RESERVED4 byte = 0x83
	RESERVED5 byte = 0x84
)

type Version byte

const (
	VERSION_1 Version = 0x01 //nolint:stylecheck,golint
)
