package goridge

import (
	"bytes"
	"encoding/gob"
	"io"
	"net/rpc"
	"sync"

	"github.com/spiral/errors"
)

// Codec represent net/rpc bridge over Goridge socket relay.
type Codec struct {
	relay  Relay
	closed bool
	frame  *Frame
	sync.Mutex
}

// NewCodec initiates new server rpc codec over socket connection.
func NewCodec(rwc io.ReadWriteCloser) *Codec {
	return &Codec{relay: NewSocketRelay(rwc)}
}

// NewCodecWithRelay initiates new server rpc codec with a relay of choice.
func NewCodecWithRelay(relay Relay) *Codec {
	return &Codec{relay: relay}
}

// ReadRequestHeader receives
func (c *Codec) ReadRequestHeader(r *rpc.Request) error {
	//c.Lock()
	//defer c.Unlock()

	frame := NewFrame()
	err := c.relay.Receive(frame)
	if err != nil {
		return err
	}

	// opts[0] sequence ID
	// opts[1] service method name offset from payload in bytes
	opts := frame.ReadOptions()
	if len(opts) != 2 {
		panic("should be 2")
	}

	r.Seq = uint64(opts[0])
	r.ServiceMethod = string(frame.Payload()[0:opts[1]])

	// save frame
	c.frame = frame

	return nil
}

// ReadRequestBody fetches prefixed body data and automatically unmarshal it as json. RawBody flag will populate
// []byte lice argument for rpc method.
func (c *Codec) ReadRequestBody(out interface{}) error {
	//c.Lock()
	//defer c.Unlock()

	frame := NewFrame()
	err := c.relay.Receive(frame)
	if err != nil {
		return err
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

	return dec.Decode(out)
}

// WriteResponse marshals response, byte slice or error to remote party.
func (c *Codec) WriteResponse(r *rpc.Response, body interface{}) error {
	const op = errors.Op("codec WriteResponse")
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

	return c.relay.Send(frame)
}

// Close underlying socket.
func (c *Codec) Close() error {
	if c.closed {
		return nil
	}

	c.closed = true
	return c.relay.Close()
}
