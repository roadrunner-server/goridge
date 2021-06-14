package rpc

import (
	"bytes"
	"io"
	"net/rpc"
	"sync"

	"github.com/spiral/errors"
	"github.com/spiral/goridge/v3/pkg/frame"
	"github.com/spiral/goridge/v3/pkg/relay"
	"github.com/spiral/goridge/v3/pkg/socket"
)

// Codec represent net/rpc bridge over Goridge socket relay.
type Codec struct {
	relay  relay.Relay
	closed bool
	frame  *frame.Frame
	codec  sync.Map

	bPool sync.Pool
	fPool sync.Pool
}

// NewCodec initiates new server rpc codec over socket connection.
func NewCodec(rwc io.ReadWriteCloser) *Codec {
	return &Codec{
		relay: socket.NewSocketRelay(rwc),
		codec: sync.Map{},

		bPool: sync.Pool{New: func() interface{} {
			return new(bytes.Buffer)
		}},

		fPool: sync.Pool{New: func() interface{} {
			return frame.NewFrame()
		}},
	}
}

// NewCodecWithRelay initiates new server rpc codec with a relay of choice.
func NewCodecWithRelay(relay relay.Relay) *Codec {
	return &Codec{relay: relay}
}

func (c *Codec) get() *bytes.Buffer {
	return c.bPool.Get().(*bytes.Buffer)
}

func (c *Codec) put(b *bytes.Buffer) {
	b.Reset()
	c.bPool.Put(b)
}

func (c *Codec) getFrame() *frame.Frame {
	return c.fPool.Get().(*frame.Frame)
}

func (c *Codec) putFrame(f *frame.Frame) {
	f.Reset()
	c.fPool.Put(f)
}

// WriteResponse marshals response, byte slice or error to remote party.
func (c *Codec) WriteResponse(r *rpc.Response, body interface{}) error { //nolint:funlen
	const op = errors.Op("goridge_write_response")
	fr := c.getFrame()
	defer c.putFrame(fr)

	// SEQ_ID + METHOD_NAME_LEN
	fr.WriteOptions(uint32(r.Seq), uint32(len(r.ServiceMethod)))
	// Write protocol version
	fr.WriteVersion(frame.VERSION_1)

	// load and delete associated codec to not waste memory
	// because we write it to the fr and don't need more information about it
	// as for go.14, Load and Delete are separate methods
	codec, ok := c.codec.Load(r.Seq)
	if !ok {
		// fallback codec
		fr.WriteFlags(frame.CODEC_GOB)
	} else {
		fr.WriteFlags(codec.(byte))
	}

	// delete the key
	c.codec.Delete(r.Seq)

	// initialize buffer
	buf := c.get()
	defer c.put(buf)

	// writeServiceMethod to the buffer
	buf.WriteString(r.ServiceMethod)

	// if error returned, we sending it via relay and return error from WriteResponse
	if r.Error != "" {
		// Append error flag
		return c.handleError(r, fr, r.Error)
	}

	// read flag previously written
	// TODO might be better to save it to local variable
	flags := fr.ReadFlags()

	switch {
	case flags&frame.CODEC_RAW != 0:
		err := encodeRaw(buf, body)
		if err != nil {
			return c.handleError(r, fr, err.Error())
		}
		// send buffer
		return c.sendBuf(fr, buf)
	case flags&frame.CODEC_PROTO != 0:
		err := encodeProto(buf, body)
		if err != nil {
			return c.handleError(r, fr, err.Error())
		}
		// send buffer
		return c.sendBuf(fr, buf)
	case flags&frame.CODEC_JSON != 0:
		err := encodeJSON(buf, body)
		if err != nil {
			return c.handleError(r, fr, err.Error())
		}
		// send buffer
		return c.sendBuf(fr, buf)
	case flags&frame.CODEC_MSGPACK != 0:
		err := encodeMsgPack(buf, body)
		if err != nil {
			return c.handleError(r, fr, err.Error())
		}
		// send buffer
		return c.sendBuf(fr, buf)
	case flags&frame.CODEC_GOB != 0:
		err := encodeGob(buf, body)
		if err != nil {
			return c.handleError(r, fr, err.Error())
		}
		// send buffer
		return c.sendBuf(fr, buf)
	default:
		return c.handleError(r, fr, errors.E(op, errors.Str("unknown codec")).Error())
	}
}

//go:inline
func (c *Codec) sendBuf(frame *frame.Frame, buf *bytes.Buffer) error {
	frame.WritePayloadLen(uint32(buf.Len()))
	frame.WritePayload(buf.Bytes())

	frame.WriteCRC()
	return c.relay.Send(frame)
}

func (c *Codec) handleError(r *rpc.Response, fr *frame.Frame, err string) error {
	buf := c.get()
	defer c.put(buf)

	// write all possible errors
	buf.WriteString(r.ServiceMethod)

	const op = errors.Op("handle codec error")
	fr.WriteFlags(frame.ERROR)
	// error should be here
	if err != "" {
		buf.WriteString(err)
	}
	fr.WritePayloadLen(uint32(buf.Len()))
	fr.WritePayload(buf.Bytes())

	fr.WriteCRC()
	_ = c.relay.Send(fr)
	return errors.E(op, errors.Str(r.Error))
}

// ReadRequestHeader receives frame with options
// options should have 2 values
// [0] - integer, sequence ID
// [1] - integer, offset for method name
// For example:
// 15Test.Payload
// SEQ_ID: 15
// METHOD_LEN: 12 and we take 12 bytes from the payload as method name
func (c *Codec) ReadRequestHeader(r *rpc.Request) error {
	const op = errors.Op("goridge_read_request_header")
	f := c.getFrame()

	err := c.relay.Receive(f)
	if err != nil {
		return err
	}

	// opts[0] sequence ID
	// opts[1] service method name offset from payload in bytes
	opts := f.ReadOptions()
	if len(opts) != 2 {
		return errors.E(op, errors.Str("should be 2 options. SEQ_ID and METHOD_LEN"))
	}

	r.Seq = uint64(opts[0])
	r.ServiceMethod = string(f.Payload()[:opts[1]])
	c.frame = f
	return c.storeCodec(r, f.ReadFlags())
}

//go:inline
func (c *Codec) storeCodec(r *rpc.Request, flag byte) error {
	switch {
	case flag&frame.CODEC_PROTO != 0:
		c.codec.Store(r.Seq, frame.CODEC_PROTO)
	case flag&frame.CODEC_JSON != 0:
		c.codec.Store(r.Seq, frame.CODEC_JSON)
	case flag&frame.CODEC_RAW != 0:
		c.codec.Store(r.Seq, frame.CODEC_RAW)
	case flag&frame.CODEC_MSGPACK != 0:
		c.codec.Store(r.Seq, frame.CODEC_MSGPACK)
	case flag&frame.CODEC_GOB != 0:
		c.codec.Store(r.Seq, frame.CODEC_GOB)
	default:
		c.codec.Store(r.Seq, frame.CODEC_GOB)
	}

	return nil
}

// ReadRequestBody fetches prefixed body data and automatically unmarshal it as json. RawBody flag will populate
// []byte lice argument for rpc method.
func (c *Codec) ReadRequestBody(out interface{}) error {
	const op = errors.Op("goridge_read_request_body")
	if out == nil {
		return nil
	}

	defer c.putFrame(c.frame)

	flags := c.frame.ReadFlags()

	switch {
	case flags&frame.CODEC_PROTO != 0:
		return decodeProto(out, c.frame)
	case flags&frame.CODEC_JSON != 0:
		return decodeJSON(out, c.frame)
	case flags&frame.CODEC_GOB != 0:
		return decodeGob(out, c.frame)
	case flags&frame.CODEC_RAW != 0:
		return decodeRaw(out, c.frame)
	case flags&frame.CODEC_MSGPACK != 0:
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
