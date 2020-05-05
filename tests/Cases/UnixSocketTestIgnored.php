<?php

/**
 * Dead simple, high performance, drop-in bridge to Golang RPC with zero dependencies
 *
 * @author Wolfy-J
 */

declare(strict_types=1);

namespace Spiral\Tests;

use Spiral\Goridge\SocketRelay;

class UnixSocketTestIgnored extends RPCTest
{
    public const SOCK_ADDR = 'server.sock';
    public const SOCK_TYPE = SocketRelay::SOCK_UNIX;
}
