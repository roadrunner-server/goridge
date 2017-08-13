package main

import (
	"fmt"
	"github.com/spiral/goridge"
	"log"
	"net"
	"net/rpc"
)

type App struct{}

func (a *App) Hi(name string, r *string) error {
	*r = fmt.Sprintf("Hello, %s!", name)
	return nil
}

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:6001")
	if err != nil {
		panic(err)
	}

	rpc.Register(new(App))
	log.Printf("started")

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		log.Printf("new connection %+v", conn)
		go rpc.ServeCodec(goridge.NewJSONCodec(conn))
	}
}
