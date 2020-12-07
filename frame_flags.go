package goridge

type FrameFlag byte

const (
	CONTEXT_SEPARATOR FrameFlag = 0x01 //nolint:golint
	PAYLOAD_CONTROL   FrameFlag = 0x02 //nolint:golint
	PAYLOAD_ERROR     FrameFlag = 0x03 //nolint:golint
)

type Version byte

const (
	VERSION_1 Version = 0x01 //nolint:golint
)
