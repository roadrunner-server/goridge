package rpc

import (
	"bytes"
	"encoding/gob"
	"io"
	"net/rpc"
	"sync"

	"github.com/goccy/go-json"
	"github.com/spiral/errors"
	"github.com/spiral/goridge/v3/pkg/frame"
	"github.com/spiral/goridge/v3/pkg/relay"
	"github.com/spiral/goridge/v3/pkg/socket"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/proto"
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
	fr.WriteOptions(fr.HeaderPtr(), uint32(r.Seq), uint32(len(r.ServiceMethod)))
	// Write protocol version
	fr.WriteVersion(fr.Header(), frame.VERSION_1)

	// load and delete associated codec to not waste memory
	// because we write it to the fr and don't need more information about it
	codec, ok := c.codec.LoadAndDelete(r.Seq)
	if !ok {
		// fallback codec
		fr.WriteFlags(fr.Header(), frame.CODEC_GOB)
	} else {
		fr.WriteFlags(fr.Header(), codec.(byte))
	}

	// if error returned, we sending it via relay and return error from WriteResponse
	if r.Error != "" {
		// Append error flag
		return c.handleError(r, fr, r.Error)
	}

	switch {
	case codec.(byte)&frame.CODEC_PROTO != 0:
		d, err := proto.Marshal(body.(proto.Message))
		if err != nil {
			return c.handleError(r, fr, err.Error())
		}

		// initialize buffer
		buf := c.get()
		defer c.put(buf)

		buf.Grow(len(d) + len(r.ServiceMethod))
		// writeServiceMethod to the buffer
		buf.WriteString(r.ServiceMethod)
		buf.Write(d)

		fr.WritePayloadLen(fr.Header(), uint32(buf.Len()))
		// copy inside
		fr.WritePayload(buf.Bytes())
		fr.WriteCRC(fr.Header())
		// send buffer
		return c.relay.Send(fr)
	case codec.(byte)&frame.CODEC_RAW != 0:
		// initialize buffer
		buf := c.get()
		defer c.put(buf)

		switch data := body.(type) {
		case []byte:
			buf.Grow(len(data) + len(r.ServiceMethod))
			// writeServiceMethod to the buffer
			buf.WriteString(r.ServiceMethod)
			buf.Write(data)

			c.frame.WritePayloadLen(c.frame.Header(), uint32(buf.Len()))
			c.frame.WritePayload(buf.Bytes())
		case *[]byte:
			buf.Grow(len(*data) + len(r.ServiceMethod))
			// writeServiceMethod to the buffer
			buf.WriteString(r.ServiceMethod)
			buf.Write(*data)

			c.frame.WritePayloadLen(c.frame.Header(), uint32(buf.Len()))
			c.frame.WritePayload(buf.Bytes())
		default:
			return c.handleError(r, fr, "unknown Raw payload type")
		}

		// send buffer
		c.frame.WriteCRC(c.frame.Header())
		return c.relay.Send(c.frame)

	case codec.(byte)&frame.CODEC_JSON != 0:
		data, err := json.Marshal(body)
		if err != nil {
			return c.handleError(r, fr, err.Error())
		}

		// initialize buffer
		buf := c.get()
		defer c.put(buf)

		buf.Grow(len(data) + len(r.ServiceMethod))
		// writeServiceMethod to the buffer
		buf.WriteString(r.ServiceMethod)
		buf.Write(data)

		fr.WritePayloadLen(fr.Header(), uint32(buf.Len()))
		// copy inside
		fr.WritePayload(buf.Bytes())
		fr.WriteCRC(fr.Header())
		// send buffer
		return c.relay.Send(fr)

	case codec.(byte)&frame.CODEC_MSGPACK != 0:
		b, err := msgpack.Marshal(body)
		if err != nil {
			return errors.E(op, err)
		}
		// initialize buffer
		buf := c.get()
		defer c.put(buf)

		buf.Grow(len(b) + len(r.ServiceMethod))
		// writeServiceMethod to the buffer
		buf.WriteString(r.ServiceMethod)
		buf.Write(b)

		fr.WritePayloadLen(fr.Header(), uint32(buf.Len()))
		// copy inside
		fr.WritePayload(buf.Bytes())
		fr.WriteCRC(fr.Header())
		// send buffer
		return c.relay.Send(fr)

	case codec.(byte)&frame.CODEC_GOB != 0:
		// initialize buffer
		buf := c.get()
		defer c.put(buf)

		buf.WriteString(r.ServiceMethod)

		dec := gob.NewEncoder(buf)
		err := dec.Encode(body)
		if err != nil {
			return errors.E(op, err)
		}

		fr.WritePayloadLen(fr.Header(), uint32(buf.Len()))
		// copy inside
		fr.WritePayload(buf.Bytes())
		fr.WriteCRC(fr.Header())
		// send buffer
		return c.relay.Send(fr)
	default:
		return c.handleError(r, fr, errors.E(op, errors.Str("unknown codec")).Error())
	}
}

func (c *Codec) handleError(r *rpc.Response, fr *frame.Frame, err string) error {
	buf := c.get()
	defer c.put(buf)

	// write all possible errors
	buf.WriteString(r.ServiceMethod)

	const op = errors.Op("handle codec error")
	fr.WriteFlags(fr.Header(), frame.ERROR)
	// error should be here
	if err != "" {
		buf.WriteString(err)
	}
	fr.WritePayloadLen(fr.Header(), uint32(buf.Len()))
	fr.WritePayload(buf.Bytes())

	fr.WriteCRC(fr.Header())
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
	opts := f.ReadOptions(f.Header())
	if len(opts) != 2 {
		return errors.E(op, errors.Str("should be 2 options. SEQ_ID and METHOD_LEN"))
	}

	r.Seq = uint64(opts[0])
	r.ServiceMethod = string(f.Payload()[:opts[1]])
	c.frame = f
	return c.storeCodec(r, f.ReadFlags())
}

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

	switch { //nolint:dupl
	case flags&frame.CODEC_PROTO != 0:
		opts := c.frame.ReadOptions(c.frame.Header())
		if len(opts) != 2 {
			return errors.E(op, errors.Str("should be 2 options. SEQ_ID and METHOD_LEN"))
		}
		payload := c.frame.Payload()[opts[1]:]
		if len(payload) == 0 {
			return nil
		}

		// check if the out message is a correct proto.Message
		// instead send an error
		if pOut, ok := out.(proto.Message); ok {
			err := proto.Unmarshal(payload, pOut)
			if err != nil {
				return errors.E(op, err)
			}
			return nil
		}

		return errors.E(op, errors.Str("message type is not a proto"))
	case flags&frame.CODEC_JSON != 0:
		opts := c.frame.ReadOptions(c.frame.Header())
		if len(opts) != 2 {
			return errors.E(op, errors.Str("should be 2 options. SEQ_ID and METHOD_LEN"))
		}
		payload := c.frame.Payload()[opts[1]:]
		if len(payload) == 0 {
			return nil
		}
		return json.Unmarshal(payload, out)
	case flags&frame.CODEC_GOB != 0:
		opts := c.frame.ReadOptions(c.frame.Header())
		if len(opts) != 2 {
			return errors.E(op, errors.Str("should be 2 options. SEQ_ID and METHOD_LEN"))
		}
		payload := c.frame.Payload()[opts[1]:]
		if len(payload) == 0 {
			return nil
		}

		buf := c.get()
		defer c.put(buf)

		dec := gob.NewDecoder(buf)
		buf.Write(payload)

		err := dec.Decode(out)
		if err != nil {
			return errors.E(op, err)
		}

		return nil
	case flags&frame.CODEC_RAW != 0:
		opts := c.frame.ReadOptions(c.frame.Header())
		if len(opts) != 2 {
			return errors.E(op, errors.Str("should be 2 options. SEQ_ID and METHOD_LEN"))
		}
		payload := c.frame.Payload()[opts[1]:]
		if len(payload) == 0 {
			return nil
		}

		if raw, ok := out.(*[]byte); ok {
			*raw = append(*raw, payload...)
		}

		return nil
	case flags&frame.CODEC_MSGPACK != 0:
		opts := c.frame.ReadOptions(c.frame.Header())
		if len(opts) != 2 {
			return errors.E(op, errors.Str("should be 2 options. SEQ_ID and METHOD_LEN"))
		}
		payload := c.frame.Payload()[opts[1]:]
		if len(payload) == 0 {
			return nil
		}

		return msgpack.Unmarshal(payload, out)
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
