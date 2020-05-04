<?php
/**
 * Dead simple, high performance, drop-in bridge to Golang RPC with zero dependencies
 *
 * @author Wolfy-J
 */

namespace Spiral\Tests\Legacy;

use Spiral\Goridge\SocketRelay;

class TCPSocketTest extends RPCTest
{
    const SOCK_ADDR = '127.0.0.1';
    const SOCK_PORT = 7079;
    const SOCK_TYPE = SocketRelay::SOCK_TCP;
}
