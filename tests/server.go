package main

import (
	"github.com/spiral/goridge"
	"net"
	"net/rpc"
	"os"
	"strings"
)

type app struct{}

type payload struct {
	Name  string            `json:"name"`
	Value int               `json:"value"`
	Keys  map[string]string `json:"keys,omitempty"`
}

func (s *app) Negate(i int64, r *int64) error {
	*r = -i
	return nil
}

func (s *app) Ping(msg string, r *string) error {
	if msg == "ping" {
		*r = "pong"
	}
	return nil
}

func (s *app) Echo(msg string, r *string) error {
	*r = msg
	return nil
}

func (s *app) Process(msg payload, r *payload) error {
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

func (s *app) EchoBinary(msg []byte, out *[]byte) error {
	*out = append(*out, msg...)

	return nil
}

func main() {
	var ln net.Listener
	var err error
	if len(os.Args) == 2 {
		ln, err = net.Listen("unix", os.Args[1])
	} else {
		ln, err = net.Listen("tcp", ":7079")
	}

	if err != nil {
		panic(err)
	}

	rpc.Register(new(app))

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeCodec(goridge.NewCodec(conn))
	}
}
