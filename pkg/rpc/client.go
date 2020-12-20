package rpc

import (
	"bytes"
	"encoding/gob"
	"io"
	"net/rpc"

	"github.com/spiral/errors"
	relay2 "github.com/spiral/goridge/v3/interfaces/relay"
	"github.com/spiral/goridge/v3/pkg/frame"
	"github.com/spiral/goridge/v3/pkg/socket"
)

// ClientCodec is codec for goridge connection.
type ClientCodec struct {
	relay  relay2.Relay
	closed bool
	frame  *frame.Frame
}

// NewClientCodec initiates new server rpc codec over socket connection.
func NewClientCodec(rwc io.ReadWriteCloser) *ClientCodec {
	return &ClientCodec{relay: socket.NewSocketRelay(rwc)}
}

// WriteRequest writes request to the connection. Sequential.
func (c *ClientCodec) WriteRequest(r *rpc.Request, body interface{}) error {
	const op = errors.Op("client codec WriteRequest")
	fr := frame.NewFrame()
	defer func() {
		// reset the fr
		fr = nil
	}()

	// for golang clients use GOB
	fr.WriteFlags(frame.CODEC_GOB)

	// SEQ_ID + METHOD_NAME_LEN
	fr.WriteOptions(uint32(r.Seq), uint32(len(r.ServiceMethod)))
	fr.WriteVersion(frame.VERSION_1)

	buf := new(bytes.Buffer)
	// writeServiceMethod to the buffer
	buf.WriteString(r.ServiceMethod)
	// Initialize gob
	if body != nil {
		enc := gob.NewEncoder(buf)
		// write data to the gob
		err := enc.Encode(body)
		if err != nil {
			return errors.E(op, err)
		}
	}

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
	const op = errors.Op("client codec ReadResponseHeader")
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
	if fr.ReadFlags()&byte(frame.ERROR) != 0 {
		r.Error = string(fr.Payload()[opts[1]:])
	}

	r.Seq = uint64(opts[0])
	r.ServiceMethod = string(fr.Payload()[:opts[1]])

	return nil
}

// ReadResponseBody response from the connection.
func (c *ClientCodec) ReadResponseBody(out interface{}) error {
	const op = errors.Op("client ReadResponseBody")
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
	case flags&byte(frame.CODEC_JSON) != 0:
		return decodeJSON(out, c.frame)
	case flags&byte(frame.CODEC_GOB) != 0:
		return decodeGob(out, c.frame)
	case flags&byte(frame.CODEC_RAW) != 0:
		return decodeRaw(out, c.frame)
	case flags&byte(frame.CODEC_MSGPACK) != 0:
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
