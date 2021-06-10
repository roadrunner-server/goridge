package rpc

import (
	"bytes"
	"encoding/gob"
	"io"
	"net/rpc"
	"sync"

	"github.com/spiral/errors"
	"github.com/spiral/goridge/v3/pkg/frame"
	"github.com/spiral/goridge/v3/pkg/relay"
	"github.com/spiral/goridge/v3/pkg/socket"
	"google.golang.org/protobuf/proto"
)

// ClientCodec is codec for goridge connection.
type ClientCodec struct {
	// bytes sync.Pool
	bPool  sync.Pool
	relay  relay.Relay
	closed bool
	frame  *frame.Frame
}

// NewClientCodec initiates new server rpc codec over socket connection.
func NewClientCodec(rwc io.ReadWriteCloser) *ClientCodec {
	return &ClientCodec{
		bPool: sync.Pool{New: func() interface{} {
			return new(bytes.Buffer)
		}},
		relay: socket.NewSocketRelay(rwc),
	}
}

func (c *ClientCodec) get() *bytes.Buffer {
	return c.bPool.Get().(*bytes.Buffer)
}

func (c *ClientCodec) put(b *bytes.Buffer) {
	b.Reset()
	c.bPool.Put(b)
}

// WriteRequest writes request to the connection. Sequential.
func (c *ClientCodec) WriteRequest(r *rpc.Request,
	body interface{}) error {
	const op = errors.Op("goridge_write_request")
	fr := frame.NewFrame()
	defer func() {
		// reset the fr
		fr = nil
	}()

	// if body is proto message, use proto codec
	buf := c.get()
	defer c.put(buf)

	// writeServiceMethod to the buffer
	buf.WriteString(r.ServiceMethod)
	// use fallback as gob
	fr.WriteFlags(frame.CODEC_GOB)

	if body != nil {
		switch m := body.(type) {
		// check if message is PROTO
		case proto.Message:
			fr.WriteFlags(frame.CODEC_PROTO)
			b, err := proto.Marshal(m)
			if err != nil {
				return errors.E(op, err)
			}
			buf.Write(b)
		default:
			enc := gob.NewEncoder(buf)
			// write data to the gob
			err := enc.Encode(body)
			if err != nil {
				return errors.E(op, err)
			}
		}
	}

	// SEQ_ID + METHOD_NAME_LEN
	fr.WriteOptions(uint32(r.Seq), uint32(len(r.ServiceMethod)))
	fr.WriteVersion(frame.VERSION_1)

	fr.WritePayloadLen(uint32(buf.Len()))
	fr.WritePayload(buf.Bytes())
	fr.WriteCRC()

	err := c.relay.Send(fr)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// ReadResponseHeader reads response from the connection.
func (c *ClientCodec) ReadResponseHeader(r *rpc.Response) error {
	const op = errors.Op("client_read_response_header")
	fr := frame.NewFrame()
	err := c.relay.Receive(fr)
	if err != nil {
		return errors.E(op, err)
	}
	if !fr.VerifyCRC() {
		return errors.E(op, errors.Str("CRC verification failed"))
	}

	// save the fr after CRC verification
	c.frame = fr

	opts := fr.ReadOptions()
	if len(opts) != 2 {
		return errors.E(op, errors.Str("should be 2 options. SEQ_ID and METHOD_LEN"))
	}

	// check for error
	if fr.ReadFlags()&frame.ERROR != 0 {
		r.Error = string(fr.Payload()[opts[1]:])
	}

	r.Seq = uint64(opts[0])
	r.ServiceMethod = string(fr.Payload()[:opts[1]])

	return nil
}

// ReadResponseBody response from the connection.
func (c *ClientCodec) ReadResponseBody(out interface{}) error {
	const op = errors.Op("client_read_response_body")
	defer func() {
		// reset the frame
		c.frame = nil
	}()
	// if there is no out interface to unmarshall the body, skip
	if out == nil {
		return nil
	}

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

// Close closes the client connection.
func (c *ClientCodec) Close() error {
	if c.closed {
		return nil
	}

	c.closed = true
	return c.relay.Close()
}
