package main

import (
	"net"
	"net/rpc"
	"os"
	"strings"
	"sync"
	"syscall"

	"github.com/spiral/goridge/v2"
)

// Service sample
type Service struct{}

// Payload sample
type Payload struct {
	Name  string            `json:"name"`
	Value int               `json:"value"`
	Keys  map[string]string `json:"keys,omitempty"`
}

// Negate number
func (s *Service) Negate(i int64, r *int64) error {
	*r = -i
	return nil
}

// Ping pong
func (s *Service) Ping(msg string, r *string) error {
	if msg == "ping" {
		*r = "pong"
	}

	return nil
}

// Echo returns incoming message
func (s *Service) Echo(msg string, r *string) error {
	*r = msg

	return nil
}

// Process performs payload conversion
func (s *Service) Process(msg Payload, r *Payload) error {
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
func (s *Service) EchoBinary(msg []byte, r *[]byte) error {
	*r = append(*r, msg...)

	return nil
}

// Service sample
type Service2 struct{}

// Payload sample
type Payload2 struct {
	Name  string            `json:"name"`
	Value int               `json:"value"`
	Keys  map[string]string `json:"keys,omitempty"`
}

// Negate number
func (s *Service2) Negate(i int64, r *int64) error {
	*r = -i
	return nil
}

// Ping pong
func (s *Service2) Ping(msg string, r *string) error {
	if msg == "ping" {
		*r = "pong"
	}

	return nil
}

// Echo returns incoming message
func (s *Service2) Echo(msg string, r *string) error {
	*r = msg

	return nil
}

// Process performs payload conversion
func (s *Service2) Process(msg Payload, r *Payload) error {
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
func (s *Service2) EchoBinary(msg []byte, r *[]byte) error {
	*r = append(*r, msg...)

	return nil
}

func listenUnixSockets(wg *sync.WaitGroup) {
	var ln net.Listener
	var err error
	wg.Add(1)
	defer wg.Done()

	if fileExists("goridge.sock") {
		err := syscall.Unlink("goridge.sock")
		if err != nil {
			panic(err)
		}
	}

	ln, err = net.Listen("unix", "goridge.sock")

	if err != nil {
		panic(err)
	}

	err = rpc.Register(new(Service2))
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		go rpc.ServeCodec(goridge.NewCodec(conn))
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func listenTCP(wg *sync.WaitGroup) {
	var ln net.Listener
	var err error
	wg.Add(1)
	defer wg.Done()

	ln, err = net.Listen("tcp", ":7079")

	if err != nil {
		panic(err)
	}

	err = rpc.Register(new(Service))
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		go rpc.ServeCodec(goridge.NewCodec(conn))
	}
}

func main() {
	wg := &sync.WaitGroup{}
	go listenTCP(wg)
	//go listenUnixSockets(wg)
	wg.Wait()
}
