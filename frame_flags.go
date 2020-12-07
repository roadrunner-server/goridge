package goridge

type FrameFlag byte

const (
	CONTEXT_SEPARATOR FrameFlag = 0x01
	PAYLOAD_CONTROL   FrameFlag = 0x02
	PAYLOAD_ERROR     FrameFlag = 0x03
)

type Version byte

const (
	VERSION_1 Version = 0x01
)
