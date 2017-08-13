<?php
/**
 * goridge
 *
 * @author    Wolfy-J
 */

namespace Spiral\Tests;

use Spiral\Goridge\Connection;
use Spiral\Tests\Prototypes\ExchangeTest;

class ExchangeTCPTest extends ExchangeTest
{
    const SOCK_ADDR = "127.0.0.1";
    const SOCK_PORT = 7078;
    const SOCK_TYPE = Connection::SOCK_TPC;
}