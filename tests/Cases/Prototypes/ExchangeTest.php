<?php
/**
 * Dead simple, high performance, drop-in bridge to Golang RPC with zero dependencies
 *
 * @author Wolfy-J
 */

namespace Spiral\Tests\Prototypes;

use PHPUnit\Framework\TestCase;
use Spiral\Goridge\Connection;

abstract class ExchangeTest extends TestCase
{
    const GO_APP    = "socket10echo";
    const SOCK_ADDR = "";
    const SOCK_PORT = 7078;
    const SOCK_TYPE = Connection::SOCK_TPC;

    public function testPingPong()
    {
        $conn = $this->makeConnection();
        $conn->send("echo");
        $this->assertSame("echo", $conn->receiveSync());
    }

    public function testPingSyncPong()
    {
        $conn = $this->makeConnection();

        $conn->send("echo1");
        $conn->send("echo2");

        $this->assertSame("echo1", $conn->receiveSync());
        $this->assertSame("echo2", $conn->receiveSync());
    }

    public function testCarryFlag()
    {
        $conn = $this->makeConnection();

        $conn->send("data", 45);

        $this->assertSame("data", $conn->receiveSync($f));
        $this->assertSame(45, $f);
    }

    /**
     * @return \Spiral\Goridge\Connection
     */
    protected function makeConnection(): Connection
    {
        return new Connection(
            static::SOCK_ADDR,
            static::SOCK_PORT,
            static::SOCK_TYPE
        );
    }
}