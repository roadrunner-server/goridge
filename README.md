High-performance PHP-to-Golang IPC bridge
=================================================
[![Latest Stable Version](https://poser.pugx.org/spiral/goridge/v/stable)](https://packagist.org/packages/spiral/goridge) 
[![GoDoc](https://godoc.org/github.com/spiral/goridge?status.svg)](https://godoc.org/github.com/spiral/goridge)
![CI](https://github.com/spiral/goridge/workflows/CI/badge.svg)
[![Scrutinizer Code Quality](https://scrutinizer-ci.com/g/spiral/goridge/badges/quality-score.png)](https://scrutinizer-ci.com/g/spiral/goridge/?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/spiral/goridge)](https://goreportcard.com/report/github.com/spiral/goridge)
[![Codecov](https://codecov.io/gh/spiral/goridge/branch/master/graph/badge.svg)](https://codecov.io/gh/spiral/goridge/)
<a href="https://discord.gg/TFeEmCs"><img src="https://img.shields.io/badge/discord-chat-magenta.svg"></a>

<img src="https://files.phpclasses.org/graphics/phpclasses/innovation-award-logo.png" height="90px" alt="PHPClasses Innovation Award" align="left"/>

Goridge is high performance PHP-to-Golang codec library which works over native PHP sockets and Golang net/rpc package.
 The library allows you to call Go service methods from PHP with minimal footprint, structures and `[]byte` support.

<br/>
See https://github.com/spiral/roadrunner - High-performance PHP application server, load-balancer and process manager written in Golang
<br/>

Features
--------
 - no external dependencies or services, drop-in (64bit PHP version required)
 - low message footprint (17 bytes over any binary payload), binary error detection
 - sockets over TCP or Unix (ext-sockets is required), standard pipes
 - very fast (300k calls per second on Ryzen 1700X over 20 threads)
 - native `net/rpc` integration, ability to connect to existed application(s)
 - standalone protocol usage
 - structured data transfer using json
 - `[]byte` transfer, including big payloads
 - service, message and transport level error handling
 - hackable
 - works on Windows

Installation
------------
```
$ go get "github.com/spiral/goridge"
```
```
$ composer require spiral/goridge
```

Example
--------
```php
<?php
use Spiral\Goridge;
require "vendor/autoload.php";

$rpc = new Goridge\RPC(new Goridge\SocketRelay("127.0.0.1", 6001));

echo $rpc->call("App.Hi", "Antony");
```

```go
package main

import (
	"fmt"
	"github.com/spiral/goridge/v2"
	"net"
	"net/rpc"
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

	rpc.Register(new(App))

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeCodec(goridge.NewCodec(conn))
	}
}
```

Check this libraries in order to find suitable socket manager:
 * https://github.com/fatih/pool
 * https://github.com/hashicorp/yamux
 
License
-------

The MIT License (MIT). Please see [`LICENSE`](./LICENSE) for more information.
