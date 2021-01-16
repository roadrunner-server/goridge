package frame

type Flag byte

// BYTE flags, it means, that we can set multiply flags from this group using bitwise OR
// For example CONTEXT_SEPARATOR | CODEC_RAW
const (
	CONTROL       Flag = 0x01
	CODEC_RAW     Flag = 0x04 //nolint:stylecheck,golint
	CODEC_JSON    Flag = 0x08 //nolint:stylecheck,golint
	CODEC_MSGPACK Flag = 0x10 //nolint:stylecheck,golint
	CODEC_GOB     Flag = 0x20 //nolint:stylecheck,golint
	ERROR         Flag = 0x40 //nolint:stylecheck,golint
	RESERVED1     Flag = 0x80
)

// COMPLEX flags can't be used with Byte flags, because it's value more than 128
const (
	RESERVED2 Flag = 0x81
	RESERVED3 Flag = 0x82
	RESERVED4 Flag = 0x83
	RESERVED5 Flag = 0x84
)

type Version byte

const (
	VERSION_1 Version = 0x01 //nolint:stylecheck,golint
)
