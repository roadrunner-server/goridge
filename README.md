High-performance PHP-to-Golang IPC transport
=================================================
[![GoDoc](https://godoc.org/github.com/roadrunner-server/goridge/v4?status.svg)](https://godoc.org/github.com/roadrunner-server/goridge/v4)
![Linux](https://github.com/roadrunner-server/goridge/workflows/Linux/badge.svg)
![macOS](https://github.com/roadrunner-server/goridge/workflows/MacOS/badge.svg)
![Windows](https://github.com/roadrunner-server/goridge/workflows/Windows/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/spiral/goridge)](https://goreportcard.com/report/github.com/spiral/goridge)
[![Codecov](https://codecov.io/gh/roadrunner-server/goridge/branch/master/graph/badge.svg)](https://codecov.io/gh/roadrunner-server/goridge/)
<a href="https://discord.gg/TFeEmCs"><img src="https://img.shields.io/badge/discord-chat-magenta.svg"></a>

<img src="https://files.phpclasses.org/graphics/phpclasses/innovation-award-logo.png" height="90px" alt="PHPClasses Innovation Award" align="left"/>

Goridge is a binary frame protocol with pipe and socket transports for inter-process communication between PHP and Go (or between any two processes that speak the goridge frame format). It exposes a 12-byte CRC-checked framing header, a small `Relay` interface, and ready-made pipe and TCP/Unix-socket implementations.
PHP source code can be found in this repository: [goridge-php](https://github.com/roadrunner-php/goridge)

<br/>
See https://github.com/roadrunner-server/roadrunner - High-performance PHP application server, load-balancer and process manager written in Golang
<br/>

Features
--------

- no external dependencies, drop-in (64bit PHP required for the PHP side)
- low message footprint (12 bytes over any binary payload), binary error detection
- CRC32 header verification
- sockets over TCP or Unix domain, standard pipes
- standalone protocol usage; bring your own RPC layer or none at all
- structured data transfer (json / msgpack / proto / gob / raw via codec flags)
- `[]byte` transfer, including large payloads
- service, message and transport level error handling
- hackable
- works on Windows (named pipes, Unix-domain sockets via AF_UNIX)

Installation
------------

```bash
go get github.com/roadrunner-server/goridge/v4
```

### Sample of usage

A minimal echo server using the socket relay:

```go
package main

import (
	"log"
	"net"

	"github.com/roadrunner-server/goridge/v4/pkg/frame"
	"github.com/roadrunner-server/goridge/v4/pkg/socket"
)

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:6001")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go serve(conn)
	}
}

func serve(conn net.Conn) {
	relay := socket.NewSocketRelay(conn)
	defer func() { _ = relay.Close() }()

	for {
		in := frame.NewFrame()
		if err := relay.Receive(in); err != nil {
			return
		}

		out := frame.NewFrame()
		out.WriteVersion(out.Header(), frame.Version1)
		out.WriteFlags(out.Header(), frame.CodecRaw)
		out.WritePayload(in.Payload())
		out.WritePayloadLen(out.Header(), uint32(len(in.Payload())))
		out.WriteCRC(out.Header())

		if err := relay.Send(out); err != nil {
			return
		}
	}
}
```

For the pipe-based variant, swap `socket.NewSocketRelay(conn)` for `pipe.NewPipeRelay(in, out)` where `in` is an `io.ReadCloser` and `out` is an `io.WriteCloser` — typically `os.Stdin`/`os.Stdout` on a child process.

License
-------

The MIT License (MIT). Please see [`LICENSE`](./LICENSE) for more information.
