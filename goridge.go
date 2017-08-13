// Author Wolfy-J, 2017. License MIT

package goridge

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/rpc"
	"reflect"
	"sync"
)

const (
	ChunkSize = 655336
)

// A JSONCodec implements ServeCodec using 9 bytes data prefixes and body marshaling via
// json.
type JSONCodec struct {
	mu     sync.Mutex // concurrent write
	rwc    io.ReadWriteCloser
	prefix Prefix // next package prefix

	closed bool
}

// NewJSONCodec initiates new JSONCodec over given connection.
func NewJSONCodec(conn io.ReadWriteCloser) *JSONCodec {
	return &JSONCodec{
		rwc:    conn,
		prefix: Prefix(make([]byte, 9)),
	}
}

// ReadRequestHeader request metadata.
func (c *JSONCodec) ReadRequestHeader(r *rpc.Request) error {
	if _, err := c.rwc.Read(c.prefix); err != nil {
		return err
	}

	if !c.prefix.HasBody() {
		return nil
	}

	method := make([]byte, c.prefix.Size())
	if _, err := c.rwc.Read(method); err != nil {
		return err
	}

	r.ServiceMethod = string(method)
	return nil
}

// ReadRequestBody fetches prefixed body payload and automatically unmarshal it as json. RawBody flag will populate
// []byte lice argument for rpc method.
func (c *JSONCodec) ReadRequestBody(out interface{}) error {
	if _, err := c.rwc.Read(c.prefix); err != nil {
		return err
	}

	if !c.prefix.HasBody() {
		return nil
	}

	// more efficient vs more memory?
	body := make([]byte, c.prefix.Size())
	body = body[:0]

	buffer := make([]byte, min(uint64(cap(body)), ChunkSize))
	doneBytes := uint64(0)

	// read only prefix.Size() from socket
	for {
		n, err := c.rwc.Read(buffer)
		if err != nil {
			return err
		}

		body = append(body, buffer[:n]...)
		doneBytes += uint64(n)

		if doneBytes == c.prefix.Size() {
			break
		}
	}

	if c.prefix.Flags()&RawBody == RawBody {
		if bin, ok := out.(*[]byte); ok {
			*bin = append(*bin, body...)
			return nil
		} else {
			return errors.New("{rawData} request for " + reflect.ValueOf(out).Elem().Kind().String())
		}

		return nil
	}

	return json.Unmarshal(body, out)
}

// WriteResponse marshaled response, byte slice or error to remote party.
func (c *JSONCodec) WriteResponse(r *rpc.Response, body interface{}) error {
	if r.Error != "" {
		log.Println("rpc: goridge error response:", r.Error)
		return c.write([]byte(r.Error), CloseConnection|ErrorBody|RawBody)
	}

	if bin, ok := body.(*[]byte); ok {
		return c.write(*bin, KeepConnection|RawBody)
	}

	res, err := json.Marshal(body)
	if err != nil {
		return err
	}

	return c.write(res, KeepConnection)
}

// Close underlying socket.
func (c *JSONCodec) Close() error {
	if c.closed {
		return nil
	}

	c.closed = true
	return c.rwc.Close()
}

// Write data prefix [flag][size uint64 LE] and payload into socket.
func (c *JSONCodec) write(payload []byte, flags byte) error {
	prefix := Prefix(make([]byte, 9))
	prefix.SetFlags(flags)
	prefix.SetSize(uint64(len(payload)))

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, err := c.rwc.Write(prefix); err != nil {
		return err
	}

	if _, err := c.rwc.Write(payload); err != nil {
		return err
	}

	return nil
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
