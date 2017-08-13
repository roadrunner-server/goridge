package main

import (
	"log"
	"net"
	"os"
)

func main() {
	var ln net.Listener
	var err error
	if len(os.Args) == 2 {
		ln, err = net.Listen("unix", os.Args[1])
	} else {
		ln, err = net.Listen("tcp", ":7077")
	}

	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
		}

		log.Printf("open %+v", conn)
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	payload := make([]byte, 4)
	for {
		_, err := conn.Read(payload)
		if err != nil {
			log.Printf("%+v errored %s", conn, err)
			conn.Close()
			break
		}

		log.Printf("got '%s'", string(payload))

		switch message := string(payload); message {
		case "ping":
			_, err := conn.Write([]byte("pong"))
			if err != nil {
				log.Println(err)
			}

			log.Println("pong")
		}
	}
}
