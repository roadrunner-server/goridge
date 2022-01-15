High-performance PHP-to-Golang IPC bridge
=================================================
[![Latest Stable Version](https://poser.pugx.org/spiral/goridge/v/stable)](https://packagist.org/packages/spiral/goridge)
[![GoDoc](https://godoc.org/github.com/roadrunner-server/goridge/v3?status.svg)](https://godoc.org/github.com/roadrunner-server/goridge/v3)
![Linux](https://github.com/roadrunner-server/goridge/workflows/Linux/badge.svg)
![macOS](https://github.com/roadrunner-server/goridge/workflows/MacOS/badge.svg)
![Windows](https://github.com/roadrunner-server/goridge/workflows/Windows/badge.svg)
![Linters](https://github.com/roadrunner-server/goridge/workflows/Linters/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/spiral/goridge)](https://goreportcard.com/report/github.com/spiral/goridge)
[![Codecov](https://codecov.io/gh/spiral/goridge/branch/master/graph/badge.svg)](https://codecov.io/gh/spiral/goridge/)
<a href="https://discord.gg/TFeEmCs"><img src="https://img.shields.io/badge/discord-chat-magenta.svg"></a>

<img src="https://files.phpclasses.org/graphics/phpclasses/innovation-award-logo.png" height="90px" alt="PHPClasses Innovation Award" align="left"/>

Goridge is high performance PHP-to-Golang codec library which works over native PHP sockets and Golang net/rpc package.
The library allows you to call Go service methods from PHP with a minimal footprint, structures and `[]byte` support.  
PHP source code can be found in this repository: [goridge-php](https://github.com/roadrunner-server/goridge-php)

<br/>
See https://github.com/spiral/roadrunner - High-performance PHP application server, load-balancer and process manager written in Golang
<br/>

Features
--------

- no external dependencies or services, drop-in (64bit PHP version required)
- low message footprint (12 bytes over any binary payload), binary error detection
- CRC32 header verification
- sockets over TCP or Unix (ext-sockets is required), standard pipes
- very fast (300k calls per second on Ryzen 1700X over 20 threads)
- native `net/rpc` integration, ability to connect to existed application(s)
- standalone protocol usage
- structured data transfer using json
- `[]byte` transfer, including big payloads
- service, message and transport level error handling
- hackable
- works on Windows
- unix sockets powered (also on Windows)

Installation
------------

```go
GO111MODULE=on go get github.com/roadrunner-server/goridge/v3
```

### Sample of usage
```go
package main

import (
	"fmt"
	"net"
	"net/rpc"

	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
)

type App struct{}

func (s *App) Hi(name string, r *string) error {
	*r = fmt.Sprintf("Hello, %s!", name)
	return nil
}

func main() {
	ln, err := net.Listen("tcp", ":6001")
	if err != nil {
		panic(err)
	}

	_ = rpc.Register(new(App))

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		_ = conn
		go rpc.ServeCodec(goridgeRpc.NewCodec(conn))
	}
}
```

License
-------

The MIT License (MIT). Please see [`LICENSE`](./LICENSE) for more information.
