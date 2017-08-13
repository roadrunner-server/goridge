Goridge, high performance PHP-to-GO net/rpc codec
=================================================
[![Latest Stable Version](https://poser.pugx.org/spiral/goridge/v/stable)](https://packagist.org/packages/spiral/goridge) 
[![GoDoc](https://godoc.org/github.com/spiral/goridge?status.svg)](https://godoc.org/github.com/spiral/goridge)
[![License](https://poser.pugx.org/spiral/goridge/license)](https://packagist.org/packages/spiral/goridge) 
[![Build Status](https://travis-ci.org/spiral/goridge.svg?branch=master)](https://travis-ci.org/spiral/goridge)
[![Scrutinizer Code Quality](https://scrutinizer-ci.com/g/spiral/goridge/badges/quality-score.png)](https://scrutinizer-ci.com/g/spiral/goridge/?branch=master)
[![Coverage Status](https://coveralls.io/repos/github/spiral/goridge/badge.svg?branch=master)](https://coveralls.io/github/spiral/goridge?branch=master)

Goridge is high-performance golang/php codec library which works over native PHP sockets and Golang net/rpc package.

Features
--------
 - no external dependencies or services, drop-in
 - low message footprint (9 bytes over binary payload)
 - sockets over TPC or Unix
 - high performance
 - native net/rpc integration, ability to connect to existed application(s)
 - structured data transer using json
 - raw binary data transfer, including big payloads
 - hackable
 - works on Windows

Examples
--------
```php
<?php
use Spiral\Goridge;
require "../vendor/autoload.php";

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
