package goridge

import (
	"io"
	"net/rpc"
	"reflect"
	"encoding/json"
	"github.com/pkg/errors"
)

// Client codec for goridge connection.
type ClientCodec struct {
	relay  Relay
	closed bool
}

// NewCodec initiates new server rpc codec over socket connection.
func NewClientCodec(rwc io.ReadWriteCloser) *ClientCodec {
	return &ClientCodec{relay: NewSocketRelay(rwc)}
}

// WriteRequest writes request to the connection. Sequential.
func (c *ClientCodec) WriteRequest(r *rpc.Request, body interface{}) error {
	if err := c.relay.Send(pack(r.ServiceMethod, r.Seq), PayloadControl|PayloadRaw); err != nil {
		return err
	}

	if bin, ok := body.(*[]byte); ok {
		return c.relay.Send(*bin, PayloadRaw)
	}

	if bin, ok := body.([]byte); ok {
		return c.relay.Send(bin, PayloadRaw)
	}

	packed, err := json.Marshal(body)
	if err != nil {
		return err
	}

	return c.relay.Send(packed, 0)
}

// ReadResponseHeader reads response from the connection.
func (c *ClientCodec) ReadResponseHeader(r *rpc.Response) error {
	data, p, err := c.relay.Receive()
	if err != nil {
		return err
	}

	if !p.HasFlag(PayloadControl) {
		return errors.New("invalid rpc header, control flag is missing")
	}

	if !p.HasFlag(PayloadRaw) {
		return errors.New("rpc response header must be in {rawData}")
	}

	if !p.HasPayload() {
		return errors.New("rpc response header can't be empty")
	}

	return unpack(data, &r.ServiceMethod, &r.Seq)
}

// ReadResponseBody response from the connection.
func (c *ClientCodec) ReadResponseBody(out interface{}) error {
	data, p, err := c.relay.Receive()
	if err != nil {
		return err
	}

	if out == nil {
		// discarding
		return nil
	}

	if !p.HasPayload() {
		return nil
	}

	if p.HasFlag(PayloadError) {
		return errors.New(string(data))
	}

	if p.HasFlag(PayloadRaw) {
		if bin, ok := out.(*[]byte); ok {
			*bin = append(*bin, data...)
			return nil
		}

		return errors.New("{rawData} request for " + reflect.ValueOf(out).String())
	}

	return json.Unmarshal(data, out)
}

// Close closes the client connection.
func (c *ClientCodec) Close() error {
	if c.closed {
		return nil
	}

	c.closed = true
	return c.relay.Close()
}
