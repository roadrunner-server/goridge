High-performance PHP-to-Golang RPC bridge
=================================================
[![Latest Stable Version](https://poser.pugx.org/spiral/goridge/v/stable)](https://packagist.org/packages/spiral/goridge) 
[![GoDoc](https://godoc.org/github.com/spiral/goridge?status.svg)](https://godoc.org/github.com/spiral/goridge)
[![License](https://poser.pugx.org/spiral/goridge/license)](https://packagist.org/packages/spiral/goridge) 
[![Build Status](https://travis-ci.org/spiral/goridge.svg?branch=master)](https://travis-ci.org/spiral/goridge)
[![Scrutinizer Code Quality](https://scrutinizer-ci.com/g/spiral/goridge/badges/quality-score.png)](https://scrutinizer-ci.com/g/spiral/goridge/?branch=master)
[![Coverage Status](https://coveralls.io/repos/github/spiral/goridge/badge.svg?branch=master)](https://coveralls.io/github/spiral/goridge?branch=master)

<img src="https://files.phpclasses.org/graphics/phpclasses/innovation-award-logo.png" height="90px" alt="PHPClasses Innovation Award" align="left"/>

Goridge is high performance PHP-to-Golang codec library which works over native PHP sockets and Golang net/rpc package. The library allows you to call Go service methods from PHP with minimal footprint, structures and `[]byte` support.

<br/><br/>
Features
--------
 - no external dependencies or services, drop-in
 - low message footprint (9 bytes over any binary payload)
 - sockets over TCP or Unix
 - very fast (260k calls per second on Ryzen 1700X over 25 threads)
 - native `net/rpc` integration, ability to connect to existed application(s)
 - structured data transfer using json
 - `[]byte` transfer, including big payloads
 - service, message and transport level error handling
 - hackable
 - works on Windows

Examples
--------
```php
<?php
use Spiral\Goridge;
require "vendor/autoload.php";

$rpc = new Goridge\JsonRPC(new Goridge\Connection("127.0.0.1", 6001));

echo $rpc->call("App.Hi", "Antony");
```

```go
package main

import (
	"fmt"
	"github.com/spiral/goridge"
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
		go rpc.ServeCodec(goridge.NewJSONCodec(conn))
	}
}
```

```
$ go get "github.com/spiral/goridge"
```

> See `tests` folder for more examples.

Check this libraries in order to find suitable socket manager:
 * https://github.com/fatih/pool
 * https://github.com/hashicorp/yamux
