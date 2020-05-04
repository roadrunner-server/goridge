<?php
/**
 * Dead simple, high performance, drop-in bridge to Golang RPC with zero dependencies
 *
 * @author Wolfy-J
 */

namespace Spiral\Tests\Legacy;

use Spiral\Goridge\RPC;
use Spiral\Goridge\SocketRelay;

class TCPSocketTest extends \Spiral\Tests\TCPSocketTest
{
    protected function makeRPC(): RPC
    {
        return new RPC(new Relay(new SocketRelay(static::SOCK_ADDR, static::SOCK_PORT, static::SOCK_TYPE)));
    }
}
