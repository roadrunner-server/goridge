<?php
/**
 * Dead simple, high performance, drop-in bridge to Golang RPC with zero dependencies
 *
 * @author Wolfy-J
 */

namespace Spiral\Tests;

use PHPUnit\Framework\TestCase;
use Spiral\Goridge\Connection;

class ConnectionTest extends TestCase
{
    public function testTcpProperties()
    {
        $conn = new Connection('localhost', 7700);

        $this->assertSame('localhost', $conn->getAddress());
        $this->assertSame(7700, $conn->getPort());
        $this->assertSame(Connection::SOCK_TPC, $conn->getType());

        $this->assertFalse($conn->isConnected());
    }

    /**
     * @expectedException \Spiral\Goridge\Exceptions\InvalidArgumentException
     * @expectedExceptionMessage no port given for TPC socket on 'localhost'
     */
    public function testTcpPortException()
    {
        $conn = new Connection('localhost', null);
    }

    /**
     * @expectedException \Spiral\Goridge\Exceptions\InvalidArgumentException
     * @expectedExceptionMessage undefined connection type 3 on 'localhost'
     */
    public function testInvalidType()
    {
        $conn = new Connection('localhost', null, 3);
    }

    public function testUnixProperties()
    {
        $conn = new Connection('/var/sock.unix', null, Connection::SOCK_UNIX);

        $this->assertSame('/var/sock.unix', $conn->getAddress());
        $this->assertSame(null, $conn->getPort());
        $this->assertSame(Connection::SOCK_UNIX, $conn->getType());

        $this->assertFalse($conn->isConnected());
    }

    public function testUnixNoPort()
    {
        $conn = new Connection('/var/sock.unix', 7700, Connection::SOCK_UNIX);
        $this->assertSame(null, $conn->getPort());
    }

    public function testToString()
    {
        $conn = new Connection('rpc.sock', null, Connection::SOCK_UNIX);
        $this->assertSame("unix://rpc.sock", (string)$conn);

        $conn = new Connection('localhost', 7700, Connection::SOCK_TPC);
        $this->assertSame("tcp://localhost:7700", (string)$conn);
    }

    /**
     * @expectedException \Spiral\Goridge\Exceptions\TransportException
     * @expectedExceptionMessage unable to close socket 'unix://rpc.sock', socket already closed
     */
    public function testCloseDeadSocket()
    {
        $conn = new Connection('rpc.sock', null, Connection::SOCK_UNIX);
        $conn->close();
    }
}