<?php
/**
 * Dead simple, high performance, drop-in bridge to Golang RPC with zero dependencies
 *
 * @author Wolfy-J
 */

namespace Spiral\Tests;

use Spiral\Goridge\Connection;
use Spiral\Tests\Prototypes\SocketTest;

class SocketTCPTest extends SocketTest
{
    const SOCK_ADDR = "127.0.0.1";
    const SOCK_PORT = 7077;
    const SOCK_TYPE = Connection::SOCK_TPC;
}