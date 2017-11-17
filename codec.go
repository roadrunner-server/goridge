package goridge

import (
	"encoding/json"
	"errors"
	"io"
	"net/rpc"
	"reflect"
)

// Codec represent net/rpc bridge over Goridge socket relay.
type Codec struct {
	relay  Relay
	closed bool
}

// NewCodec initiates new server rpc codec over socket connection.
func NewCodec(rwc io.ReadWriteCloser) *Codec {
	return &Codec{relay: NewSocketRelay(rwc)}
}

// ReadRequestHeader receives
func (c *Codec) ReadRequestHeader(r *rpc.Request) error {
	data, p, err := c.relay.Receive()
	if err != nil {
		return err
	}

	if !p.HasFlag(PayloadControl) {
		return errors.New("invalid request, control data is expected")
	}

	if !p.HasFlag(PayloadRaw) {
		return errors.New("rpc control command must be in {rawData}")
	}

	if !p.HasPayload() {
		return nil
	}

	r.ServiceMethod = string(data)
	return nil
}

// ReadRequestBody fetches prefixed body data and automatically unmarshal it as json. RawBody flag will populate
// []byte lice argument for rpc method.
func (c *Codec) ReadRequestBody(out interface{}) error {
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

// WriteResponse marshals response, byte slice or error to remote party.
func (c *Codec) WriteResponse(r *rpc.Response, body interface{}) error {
	if r.Error != "" {
		return c.relay.Send([]byte(r.Error), PayloadError|PayloadRaw)
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

// Close underlying socket.
func (c *Codec) Close() error {
	if c.closed {
		return nil
	}

	c.closed = true
	return c.relay.Close()
}
