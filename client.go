package goridge

import (
	"bytes"
	"encoding/gob"
	"io"
	"net/rpc"
	"sync"

	"github.com/spiral/errors"
)

// ClientCodec is codec for goridge connection.
type ClientCodec struct {
	relay  Relay
	closed bool
	frame  *Frame
	sync.Mutex
}

// NewClientCodec initiates new server rpc codec over socket connection.
func NewClientCodec(rwc io.ReadWriteCloser) *ClientCodec {
	return &ClientCodec{relay: NewSocketRelay(rwc)}
}

// WriteRequest writes request to the connection. Sequential.
func (c *ClientCodec) WriteRequest(r *rpc.Request, body interface{}) error {
	const op = errors.Op("client codec WriteRequest")

	frame := NewFrame()
	frame.WriteFlags(CONTEXT_SEPARATOR, CODEC_GOB)
	// SEQ_ID + METHOD_NAME_LEN
	frame.WriteOptions(uint32(r.Seq), uint32(len(r.ServiceMethod)))
	frame.WriteVersion(VERSION_1)

	buf := new(bytes.Buffer)
	// writeServiceMethod to the buffer
	buf.WriteString(r.ServiceMethod)
	// Initialize gob
	enc := gob.NewEncoder(buf)
	// write data to the gob
	err := enc.Encode(body)
	if err != nil {
		return errors.E(op, err)
	}

	frame.WritePayloadLen(uint32(buf.Len()))
	frame.WritePayload(buf.Bytes())
	frame.WriteCRC()

	err = c.relay.Send(frame)
	if err != nil {
		return errors.E(op, err)
	}
	err = c.relay.Send(frame)
	if err != nil {
		return errors.E(op, err)
	}

	// reset the frame
	frame = nil

	return nil
}

// ReadResponseHeader reads response from the connection.
func (c *ClientCodec) ReadResponseHeader(r *rpc.Response) error {
	const op = errors.Op("client codec ReadResponseHeader")
	frame := NewFrame()
	err := c.relay.Receive(frame)
	if err != nil {
		return errors.E(op, err)
	}
	if !frame.VerifyCRC() {
		return errors.E(op, errors.Str("CRC verification failed"))
	}

	// save the frame after CRC verification
	c.frame = frame

	opts := frame.ReadOptions()
	if len(opts) != 2 {
		return errors.E(op, errors.Str("should be 2 options. SEQ_ID and METHOD_LEN"))
	}

	// check for error
	if frame.ReadFlags()&uint8(ERROR) != 0 {
		r.Error = string(frame.Payload()[opts[1]:])
	}

	r.Seq = uint64(opts[0])
	r.ServiceMethod = string(frame.Payload()[0:opts[1]])

	return nil
}

// ReadResponseBody response from the connection.
func (c *ClientCodec) ReadResponseBody(out interface{}) error {
	const op = errors.Op("client ReadResponseBody")
	if out == nil {
		// reset the frame
		c.frame = nil
		return nil
	}
	buf := new(bytes.Buffer)
	dec := gob.NewDecoder(buf)

	opts := c.frame.ReadOptions()
	if len(opts) != 2 {
		panic("should be 2")
	}

	payload := c.frame.Payload()[opts[1]:]
	buf.Write(payload)

	// reset the frame
	c.frame = nil

	err := dec.Decode(out)
	if err != nil {
		return errors.E(op, err)
	}
	buf.Truncate(0)
	return nil
}

// Close closes the client connection.
func (c *ClientCodec) Close() error {
	if c.closed {
		return nil
	}

	c.closed = true
	return c.relay.Close()
}
