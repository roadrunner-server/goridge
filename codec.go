package goridge

import (
	"bytes"
	"io"
	"net/rpc"
	"sync"

	j "github.com/json-iterator/go"
	"github.com/spiral/errors"
	"go.uber.org/multierr"
)

var json = j.ConfigCompatibleWithStandardLibrary

// Codec represent net/rpc bridge over Goridge socket relay.
type Codec struct {
	relay  Relay
	closed bool
	frame  *Frame
	codec  sync.Map
}

// NewCodec initiates new server rpc codec over socket connection.
func NewCodec(rwc io.ReadWriteCloser) *Codec {
	return &Codec{
		relay: NewSocketRelay(rwc),
		codec: sync.Map{},
	}
}

// NewCodecWithRelay initiates new server rpc codec with a relay of choice.
func NewCodecWithRelay(relay Relay) *Codec {
	return &Codec{relay: relay}
}

// WriteResponse marshals response, byte slice or error to remote party.
func (c *Codec) WriteResponse(r *rpc.Response, body interface{}) error {
	const op = errors.Op("codec: write response")
	frame := NewFrame()

	// SEQ_ID + METHOD_NAME_LEN
	frame.WriteOptions(uint32(r.Seq), uint32(len(r.ServiceMethod)))
	// Write protocol version
	frame.WriteVersion(VERSION_1)

	// load and delete associated codec to not waste memory
	// because we write it to the frame and don't need more information about it
	// as for go.14, Load and Delete are separate methods
	codec, ok := c.codec.Load(r.Seq)
	if !ok {
		// fallback codec
		frame.WriteFlags(CODEC_GOB)
	} else {
		frame.WriteFlags(codec.(FrameFlag))
	}

	// delete the key
	c.codec.Delete(r.Seq)

	// initialize buffer
	buf := new(bytes.Buffer)
	// writeServiceMethod to the buffer
	buf.WriteString(r.ServiceMethod)

	// if error returned, we sending it via relay and return error from WriteResponse
	if r.Error != "" {
		// Append error flag
		return c.handleError(r, frame, buf, errors.Str(r.Error))
	}

	// read flag previously written
	// TODO might be better to save it to local variable
	flags := frame.ReadFlags()

	switch {
	case flags&byte(CODEC_JSON) != 0:
		err := encodeJSON(buf, body)
		if err != nil {
			return c.handleError(r, frame, buf, err)
		}
		// send buffer
		return c.sendBuf(frame, buf)
	case flags&byte(CODEC_MSGPACK) != 0:
		err := encodeMsgPack(buf, body)
		if err != nil {
			return c.handleError(r, frame, buf, err)
		}
		// send buffer
		return c.sendBuf(frame, buf)
	case flags&byte(CODEC_RAW) != 0:
		err := encodeRaw(buf, body)
		if err != nil {
			return c.handleError(r, frame, buf, err)
		}
		// send buffer
		return c.sendBuf(frame, buf)
	case flags&byte(CODEC_GOB) != 0:
		err := encodeGob(buf, body)
		if err != nil {
			return c.handleError(r, frame, buf, err)
		}
		// send buffer
		return c.sendBuf(frame, buf)
	default:
		return c.handleError(r, frame, buf, errors.E(op, errors.Str("unknown codec")))
	}
}

func (c *Codec) sendBuf(frame *Frame, buf *bytes.Buffer) error {
	frame.WritePayloadLen(uint32(buf.Len()))
	frame.WritePayload(buf.Bytes())

	frame.WriteCRC()
	return c.relay.Send(frame)
}

func (c *Codec) handleError(r *rpc.Response, frame *Frame, buf *bytes.Buffer, err error) error {
	// just to be sure, remove all data from buffer and write new
	buf.Truncate(0)
	// write all possible errors
	err = multierr.Append(err, errors.Str(r.Error))
	buf.WriteString(r.ServiceMethod)

	const op = errors.Op("handle codec error")
	frame.WriteFlags(ERROR)
	// error should be here
	if err != nil {
		buf.WriteString(err.Error())
	}
	frame.WritePayloadLen(uint32(buf.Len()))
	frame.WritePayload(buf.Bytes())

	frame.WriteCRC()
	_ = c.relay.Send(frame)
	return errors.E(op, errors.Str(r.Error))
}

// ReadRequestHeader receives frame with options
// options should have 2 values
// [0] - integer, sequence ID
// [1] - integer, offset for method name
// For example:
// 15Test.Payload
// SEQ_ID: 15
// METHOD_LEN: 11 and we take 11bytes from the payload as method name
func (c *Codec) ReadRequestHeader(r *rpc.Request) error {
	const op = errors.Op("codec: read request header")
	frame := NewFrame()
	err := c.relay.Receive(frame)
	if err != nil {
		return err
	}

	// opts[0] sequence ID
	// opts[1] service method name offset from payload in bytes
	opts := frame.ReadOptions()
	if len(opts) != 2 {
		return errors.E(op, errors.Str("should be 2 options. SEQ_ID and METHOD_LEN"))
	}

	r.Seq = uint64(opts[0])
	r.ServiceMethod = string(frame.Payload()[:opts[1]])
	c.frame = frame
	return c.storeCodec(r, frame.ReadFlags())
}

func (c *Codec) storeCodec(r *rpc.Request, flag byte) error {
	switch {
	case flag&byte(CODEC_JSON) != 0:
		c.codec.Store(r.Seq, CODEC_JSON)
	case flag&byte(CODEC_RAW) != 0:
		c.codec.Store(r.Seq, CODEC_RAW)
	case flag&byte(CODEC_MSGPACK) != 0:
		c.codec.Store(r.Seq, CODEC_MSGPACK)
	case flag&byte(CODEC_GOB) != 0:
		c.codec.Store(r.Seq, CODEC_GOB)
	default:
		c.codec.Store(r.Seq, CODEC_GOB)
	}

	return nil
}

// ReadRequestBody fetches prefixed body data and automatically unmarshal it as json. RawBody flag will populate
// []byte lice argument for rpc method.
func (c *Codec) ReadRequestBody(out interface{}) error {
	const op = errors.Op("codec read request body")
	if out == nil {
		return nil
	}

	defer func() {
		c.frame = nil
	}()

	flags := c.frame.ReadFlags()

	switch {
	case flags&byte(CODEC_JSON) != 0:
		return decodeJSON(out, c.frame)
	case flags&byte(CODEC_GOB) != 0:
		return decodeGob(out, c.frame)
	case flags&byte(CODEC_RAW) != 0:
		return decodeRaw(out, c.frame)
	case flags&byte(CODEC_MSGPACK) != 0:
		return decodeMsgPack(out, c.frame)
	default:
		return errors.E(op, errors.Str("unknown decoder used in frame"))
	}
}

// Close underlying socket.
func (c *Codec) Close() error {
	if c.closed {
		return nil
	}

	c.closed = true
	return c.relay.Close()
}
