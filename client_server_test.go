package goridge

import (
	"net"
	"net/rpc"
	"strings"
	"sync"
	"testing"

	"github.com/spiral/errors"
	"github.com/stretchr/testify/assert"
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
	*r = "error"
	return errors.Str("echoR error")
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
func TestClientServerJSON(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:18935")
	assert.NoError(t, err)

	go func() {
		for {
			conn, err2 := ln.Accept()
			assert.NoError(t, err2)
			rpc.ServeCodec(NewCodec(conn))
		}
	}()

	err = rpc.RegisterName("test2", new(testService))
	assert.NoError(t, err)

	conn, err := net.Dial("tcp", "127.0.0.1:18935")
	assert.NoError(t, err)

	client := rpc.NewClientWithCodec(NewClientCodec(conn))
	defer func() {
		err := client.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	var rp = Payload{}
	assert.NoError(t, client.Call("test2.Process", Payload{
		Name:  "name",
		Value: 1000,
		Keys:  map[string]string{"key": "value"},
	}, &rp))

	assert.Equal(t, "NAME", rp.Name)
	assert.Equal(t, -1000, rp.Value)
	assert.Equal(t, "key", rp.Keys["value"])
}

func TestClientServerConcurrent(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:22385")
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			conn, err2 := ln.Accept()
			assert.NoError(t, err2)
			rpc.ServeCodec(NewCodec(conn))
		}
	}()

	err = rpc.RegisterName("test", new(testService))
	assert.NoError(t, err)

	conn, err := net.Dial("tcp", "127.0.0.1:22385")
	assert.NoError(t, err)

	client := rpc.NewClientWithCodec(NewClientCodec(conn))
	defer func() {
		err := client.Close()
		assert.NoError(t, err)
	}()

	wg := &sync.WaitGroup{}
	wg.Add(300)

	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			var rp = Payload{}
			d := client.Go("test.Process", Payload{
				Name:  "name",
				Value: 1000,
				Keys:  map[string]string{"key": "value"},
			}, &rp, nil)

			<-d.Done
			assert.Equal(t, "NAME", rp.Name)
			assert.Equal(t, -1000, rp.Value)
			assert.Equal(t, "key", rp.Keys["value"])
		}()

		go func() {
			defer wg.Done()
			var rs = ""
			d := client.Go("test.Echo", "hello", &rs, nil)
			<-d.Done
			assert.Equal(t, "hello", rs)
		}()

		go func() {
			defer wg.Done()
			rs := ""
			rb := make([]byte, 0)

			a := client.Go("test.Echo", "hello", &rs, nil)
			b := client.Go("test.EchoBinary", []byte("hello world"), &rb, nil)
			c := client.Go("test.EchoR", "hi", &rs, nil)

			for i := 0; i < 3; i++ {
				select {
				case reply := <-a.Done:
					_ = reply
					assert.Equal(t, "hello", rs)
				case reply := <-b.Done:
					_ = reply
					assert.Equal(t, []byte("hello world"), rb)
				case reply := <-c.Done:
					assert.Error(t, reply.Error)
				}
			}
		}()
	}

	wg.Wait()

	wg2 := &sync.WaitGroup{}
	wg2.Add(300)

	for i := 0; i < 100; i++ {
		go func() {
			defer wg2.Done()
			var rp = Payload{}
			assert.NoError(t, client.Call("test.Process", Payload{
				Name:  "name",
				Value: 1000,
				Keys:  map[string]string{"key": "value"},
			}, &rp))

			assert.Equal(t, "NAME", rp.Name)
			assert.Equal(t, -1000, rp.Value)
			assert.Equal(t, "key", rp.Keys["value"])
		}()

		go func() {
			defer wg2.Done()
			var rs = ""
			assert.NoError(t, client.Call("test.Echo", "hello", &rs))
			assert.Equal(t, "hello", rs)
		}()

		go func() {
			defer wg2.Done()
			rs := ""
			rb := make([]byte, 0, len("hello world"))
			assert.NoError(t, client.Call("test.Echo", "hello", &rs))
			assert.Equal(t, "hello", rs)

			assert.NoError(t, client.Call("test.EchoBinary", []byte("hello world"), &rb))
			assert.Equal(t, []byte("hello world"), rb)

			assert.Error(t, client.Call("test.EchoR", "hi", &rs))
		}()
	}

	wg2.Wait()
}
