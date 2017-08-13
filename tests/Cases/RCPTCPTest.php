<?php
/**
 * goridge
 *
 * @author    Wolfy-J
 */

namespace Spiral\Tests;

use Spiral\Goridge\Connection;
use Spiral\Tests\Prototypes\RPCTest;

class RCPTCPTest extends RPCTest
{
    const SOCK_ADDR = "127.0.0.1";
    const SOCK_PORT = 7079;
    const SOCK_TYPE = Connection::SOCK_TPC;
}