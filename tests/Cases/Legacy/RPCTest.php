<?php

declare(strict_types=1);

namespace Spiral\Tests\Legacy;

use Spiral\Goridge\RPC;
use Spiral\Goridge\SocketRelay;

class RPCTest extends \Spiral\Tests\RPCTest
{
    protected function makeRPC(): RPC
    {
        return new RPC(new Relay(new SocketRelay(static::SOCK_ADDR, static::SOCK_PORT, static::SOCK_TYPE)));
    }
}
