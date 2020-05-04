<?php
/**
 * Dead simple, high performance, drop-in bridge to Golang RPC with zero dependencies
 *
 * @author Wolfy-J
 */

namespace Spiral\Tests;

use Spiral\Goridge\SocketRelay;

class UnixSocketTest_ extends RPCTest
{
    const SOCK_ADDR = 'server.sock';
    const SOCK_TYPE = SocketRelay::SOCK_UNIX;
}
