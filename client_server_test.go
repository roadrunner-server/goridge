package goridge

import (
	"crypto/rand"
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

	// this test uses random inputs
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			var rp = Payload{}
			b := make([]byte, 15)
			_, err := rand.Read(b)
			assert.NoError(t, err)

			<-client.Go("test.Process", Payload{
				Name:  string(b),
				Value: 1000,
				Keys:  map[string]string{"key": string(b)},
			}, &rp, nil).Done

			assert.Equal(t, strings.ToUpper(string(b)), rp.Name)
			assert.Equal(t, -1000, rp.Value)
			assert.Equal(t, "key", rp.Keys[string(b)])
		}()

		go func() {
			var rs = ""
			b := make([]byte, 15)
			_, err := rand.Read(b)
			assert.NoError(t, err)
			<-client.Go("test.Echo", string(b), &rs, nil).Done
			assert.Equal(t, string(b), rs)
			wg.Done()
		}()

		go func() {
			rs := ""
			rb := make([]byte, 0)

			r := make([]byte, 15)
			_, err := rand.Read(r)
			assert.NoError(t, err)
			a := client.Go("test.Echo", string(r), &rs, nil)
			b := client.Go("test.EchoBinary", []byte("hello world"), &rb, nil)
			c := client.Go("test.EchoR", "hi", &rs, nil)

			<-a.Done
			assert.Equal(t, string(r), rs)
			<-b.Done
			assert.Equal(t, []byte("hello world"), rb)
			resC := <-c.Done
			assert.Error(t, resC.Error)
			wg.Done()
		}()
	}

	wg.Wait()

	wg2 := &sync.WaitGroup{}
	wg2.Add(300)

	for i := 0; i < 100; i++ {
		go func() {
			defer wg2.Done()
			var rp = Payload{}
			b := make([]byte, 15)
			_, err := rand.Read(b)
			assert.NoError(t, err)

			assert.NoError(t, client.Call("test.Process", Payload{
				Name:  string(b),
				Value: 1000,
				Keys:  map[string]string{"key": string(b)},
			}, &rp))

			assert.Equal(t, strings.ToUpper(string(b)), rp.Name)
			assert.Equal(t, -1000, rp.Value)
			assert.Equal(t, "key", rp.Keys[string(b)])
		}()

		go func() {
			defer wg2.Done()
			var rs = ""
			r := make([]byte, 15)
			_, err := rand.Read(r)
			assert.NoError(t, err)

			assert.NoError(t, client.Call("test.Echo", string(r), &rs))
			assert.Equal(t, string(r), rs)
		}()

		go func() {
			defer wg2.Done()
			rs := ""
			rb := make([]byte, 0, len("hello world"))

			r := make([]byte, 15)
			_, err := rand.Read(r)
			assert.NoError(t, err)

			assert.NoError(t, client.Call("test.Echo", string(r), &rs))
			assert.Equal(t, string(r), rs)

			assert.NoError(t, client.Call("test.EchoBinary", r, &rb))
			assert.Equal(t, r, rb)

			assert.Error(t, client.Call("test.EchoR", "hi", &rs))
		}()
	}

	wg2.Wait()
}
