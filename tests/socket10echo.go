package main

import (
	"log"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":7078")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
		}

		log.Printf("open %+v", conn)
		go handleConn2(conn)
	}
}

func handleConn2(conn net.Conn) {
	payload := make([]byte, 512)

	for {
		n, err := conn.Read(payload)
		if err != nil {
			log.Printf("%+v errored %s", conn, err)
			conn.Close()
		}

		payload = payload[:n]
		log.Printf("echo '%v' of len %v", payload, len(payload))

		n, err = conn.Write(payload)
		if err != nil {
			log.Println(err)
		}

		log.Printf("sent %v bytes of %v", n, payload)
	}
}
