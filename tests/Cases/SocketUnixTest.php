<?php
/**
 * goridge
 *
 * @author    Wolfy-J
 */

namespace Spiral\Tests;

use Spiral\Goridge\Connection;
use Spiral\Tests\Prototypes\SocketTest;

class SocketUnixTest extends SocketTest
{
    const SOCK_ADDR = "../socket4bytes.sock";
    const SOCK_TYPE = Connection::SOCK_UNIX;

    public function setUp()
    {
        if (strtoupper(substr(PHP_OS, 0, 3)) === 'WIN') {
            $this->markTestSkipped("not available at windows");

            return;
        }

        parent::setUp();
    }
}