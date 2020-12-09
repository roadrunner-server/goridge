package goridge

type FrameFlag byte

// COMPOSED FLAGS, it means, that we can set multiply flags from this group
const (
	CONTEXT_SEPARATOR FrameFlag = 0x00 //nolint:golint
	CODEC_RAW         FrameFlag = 0x01 //nolint:golint
	CODEC_JSON        FrameFlag = 0x04 //nolint:golint
	CODEC_MSGPACK     FrameFlag = 0x08 //nolint:golint
	CODEC_GOB         FrameFlag = 0x10 //nolint:golint
	ERROR             FrameFlag = 0x20 //nolint:golint
	RESERVED1         FrameFlag = 0x40 //nolint:golint
	RESERVED2         FrameFlag = 0x80 //nolint:golint
)

// SINGLE FLAGS
const ()

type Version byte

const (
	VERSION_1 Version = 0x01 //nolint:golint
)
