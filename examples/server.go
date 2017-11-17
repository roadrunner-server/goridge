package main

import (
	"fmt"
	"github.com/spiral/goridge"
	"log"
	"net"
	"net/rpc"
)

type app struct{}

func (a *app) Hi(name string, r *string) error {
	*r = fmt.Sprintf("Hello, %s!", name)
	return nil
}

func main() {
	ln, err := net.Listen("tcp", ":6001")
	if err != nil {
		panic(err)
	}

	rpc.Register(new(app))
	log.Printf("started")

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		log.Printf("new connection %+v", conn)
		go rpc.ServeCodec(goridge.NewCodec(conn))
	}
}
