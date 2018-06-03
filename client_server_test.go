package goridge

import (
	"testing"
	"strings"
	"net"
	"net/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/pkg/errors"
)

// testService sample
type testService struct{}

// Payload sample
type Payload struct {
	Name  string            `json:"name"`
	Value int               `json:"value"`
	Keys  map[string]string `json:"keys,omitempty"`
}

// Echo returns incoming message
func (s *testService) Echo(msg string, r *string) error {
	*r = msg
	return nil
}

// Echo returns error
func (s *testService) EchoR(msg string, r *string) error {
	return errors.New("echoR error")
}

// Process performs payload conversion
func (s *testService) Process(msg Payload, r *Payload) error {
	r.Name = strings.ToUpper(msg.Name)
	r.Value = -msg.Value

	if len(msg.Keys) != 0 {
		r.Keys = make(map[string]string)
		for n, v := range msg.Keys {
			r.Keys[v] = n
		}
	}

	return nil
}

// EchoBinary work over binary data
func (s *testService) EchoBinary(msg []byte, out *[]byte) error {
	*out = append(*out, msg...)

	return nil
}

func TestClientServer(t *testing.T) {
	var ln net.Listener
	var err error

	ln, err = net.Listen("tcp", ":8079")
	if err != nil {
		panic(err)
	}

	rpc.RegisterName("test", new(testService))

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				continue
			}
			rpc.ServeCodec(NewCodec(conn))
		}
	}()

	conn, err := net.Dial("tcp", ":8079")
	if err != nil {
		panic(err)
	}

	client := rpc.NewClientWithCodec(NewClientCodec(conn))
	defer client.Close()

	var (
		rs = ""
		rp = Payload{}
		rb = make([]byte, 0)
	)

	assert.NoError(t, client.Call("test.Process", Payload{
		Name:  "name",
		Value: 1000,
		Keys:  map[string]string{"key": "value"},
	}, &rp))

	assert.Equal(t, "NAME", rp.Name)
	assert.Equal(t, -1000, rp.Value)
	assert.Equal(t, "key", rp.Keys["value"])

	assert.NoError(t, client.Call("test.Echo", "hello", &rs))
	assert.Equal(t, "hello", rs)

	assert.NoError(t, client.Call("test.EchoBinary", []byte("hello world"), &rb))
	assert.Equal(t, []byte("hello world"), rb)

	assert.Error(t, client.Call("test.EchoR", "hi", &rs))
}
