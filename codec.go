package goridge

import (
	"bytes"
	"encoding/gob"
	"io"
	"net/rpc"
	"sync"

	j "github.com/json-iterator/go"
	"github.com/spiral/errors"
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
	const op = errors.Op("codec WriteResponse")
	frame := NewFrame()

	// SEQ_ID + METHOD_NAME_LEN
	frame.WriteOptions(uint32(r.Seq), uint32(len(r.ServiceMethod)))
	// Write protocol version
	frame.WriteVersion(VERSION_1)

	codec, ok := c.codec.Load(r.Seq)
	if ok == false {
		// fallback codec
		frame.WriteFlags(CODEC_GOB)
	} else {
		frame.WriteFlags(codec.(FrameFlag))
	}

	buf := new(bytes.Buffer)
	// writeServiceMethod to the buffer
	buf.WriteString(r.ServiceMethod)

	// if error returned, we sending it via relay and return error from WriteResponse
	if r.Error != "" {
		frame.WriteFlags(ERROR)
		// write data to the gob
		buf.WriteString(r.Error)

		frame.WritePayloadLen(uint32(buf.Len()))
		frame.WritePayload(buf.Bytes())

		frame.WriteCRC()
		_ = c.relay.Send(frame)
		return errors.E(op, errors.Str(r.Error))
	}

	// Initialize gob encoder
	enc := gob.NewEncoder(buf)

	// write data to the gob
	err := enc.Encode(body)
	if err != nil {
		return errors.E(op, err)
	}

	frame.WritePayloadLen(uint32(buf.Len()))
	frame.WritePayload(buf.Bytes())

	frame.WriteCRC()

	return c.relay.Send(frame)
}

// ReadRequestHeader receives
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
	if flag&byte(CODEC_JSON) != 0 {
		c.codec.Store(r.Seq, CODEC_JSON)
		return nil
	}

	if flag&byte(CODEC_GOB) != 0 {
		c.codec.Store(r.Seq, CODEC_GOB)
		return nil
	}

	if flag&byte(CODEC_RAW) != 0 {
		c.codec.Store(r.Seq, CODEC_RAW)
		return nil
	}

	if flag&byte(CODEC_MSGPACK) != 0 {
		c.codec.Store(r.Seq, CODEC_MSGPACK)
		return nil
	}

	// If no flags provided, store GOB
	c.codec.Store(r.Seq, CODEC_GOB)
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

	if flags&byte(CODEC_JSON) != byte(0) {
		return decodeJSON(out, c.frame)
	}

	if flags&byte(CODEC_GOB) != byte(0) {
		return decodeGob(out, c.frame)
	}

	if flags&byte(CODEC_RAW) != byte(0) {
		return decodeRaw(out, c.frame)
	}

	if flags&byte(CODEC_MSGPACK) != byte(0) {
		return decodeMsgPack(out, c.frame)
	}

	return errors.E(op, errors.Str("unknown decoder used in frame"))
}

// Close underlying socket.
func (c *Codec) Close() error {
	if c.closed {
		return nil
	}

	c.closed = true
	return c.relay.Close()
}
