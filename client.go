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
	if err := c.relay.Send([]byte(r.ServiceMethod), PayloadControl|PayloadRaw); err != nil {
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

// ReadResponseHeader reads response from the connection. Sequential.
func (c *ClientCodec) ReadResponseHeader(r *rpc.Response) error {
	data, p, err := c.relay.Receive()
	if err != nil {
		return err
	}

	if !p.HasFlag(PayloadControl) {
		return errors.New("invalid response, control data is expected")
	}

	if !p.HasFlag(PayloadRaw) {
		return errors.New("rpc response control command must be in {rawData}")
	}
	if !p.HasPayload() {
		return nil
	}

	r.ServiceMethod = string(data)
	return nil
}

// ReadResponseBody response from the connection.
func (c *ClientCodec) ReadResponseBody(out interface{}) error {
	data, p, err := c.relay.Receive()
	if err != nil {
		return err
	}

	if !p.HasPayload() {
		return nil
	}

	if p.HasFlag(PayloadRaw) {
		if bin, ok := out.(*[]byte); ok {
			*bin = append(*bin, data...)
			return nil
		}

		return errors.New("{rawData} request for " + reflect.ValueOf(out).Elem().Kind().String())
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
