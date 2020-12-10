package goridge

type FrameFlag byte

// BYTE flags, it means, that we can set multiply flags from this group using bitwise OR
// For example CONTEXT_SEPARATOR | CODEC_RAW
const (
	CONTEXT_SEPARATOR FrameFlag = 0x01 //nolint:stylecheck,golint
	CODEC_RAW         FrameFlag = 0x04 //nolint:stylecheck,golint
	CODEC_JSON        FrameFlag = 0x08 //nolint:stylecheck,golint
	CODEC_MSGPACK     FrameFlag = 0x10 //nolint:stylecheck,golint
	CODEC_GOB         FrameFlag = 0x20 //nolint:stylecheck,golint
	ERROR             FrameFlag = 0x40 //nolint:stylecheck,golint
	RESERVED1         FrameFlag = 0x80 //nolint:stylecheck,golint
)

// COMPLEX flags can't be used with Byte flags, because it's value more than 128
const (
	RESERVED2 FrameFlag = 0x81 //nolint:stylecheck,golint
	RESERVED3 FrameFlag = 0x82 //nolint:stylecheck,golint
	RESERVED4 FrameFlag = 0x83 //nolint:stylecheck,golint
	RESERVED5 FrameFlag = 0x84 //nolint:stylecheck,golint
)

type Version byte

const (
	VERSION_1 Version = 0x01 //nolint:stylecheck,golint
)
