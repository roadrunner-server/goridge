package goridge

type FrameFlag byte

const (
	CONTEXT_SEPARATOR FrameFlag = 0x00 //nolint:golint
	CODEC_RAW         FrameFlag = 0x01 //nolint:golint
	CODEC_JSON        FrameFlag = 0x02 //nolint:golint
	CODEC_MSGPACK     FrameFlag = 0x03 //nolint:golint
	CODEC_GOB         FrameFlag = 0x04 //nolint:golint
)

type Version byte

const (
	VERSION_1 Version = 0x01 //nolint:golint
)
