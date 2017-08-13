<?php
/**
 * goridge
 *
 * @author    Wolfy-J
 */

namespace Spiral\Tests\Prototypes;

use Spiral\Goridge\Connection;

abstract class SocketTest extends ProcessTest
{
    const GO_APP    = "socket4bytes";
    const SOCK_ADDR = "";
    const SOCK_PORT = 7077;
    const SOCK_TYPE = Connection::SOCK_TPC;

    /**
     * @expectedException \Spiral\Goridge\Exceptions\TransportException
     * @expectedExceptionMessageRegExp #unable to establish connection .*#
     */
    public function testFailToConnect()
    {
        $this->process->signal(0);
        $this->makeConnection();
    }

    public function testPingSocket()
    {
        $conn = $this->makeConnection($socket);
        $this->assertTrue($conn->isConnected());

        socket_send($socket, "ping", 4, 0);
        $this->assertSame("pong", socket_read($socket, 4, PHP_BINARY_READ));
    }

    public function testPingSocketTwice()
    {
        $conn = $this->makeConnection($socket);
        $this->assertTrue($conn->isConnected());

        socket_send($socket, "ping", 4, 0);
        $this->assertSame("pong", socket_read($socket, 4, PHP_BINARY_READ));

        socket_send($socket, "ping", 4, 0);
        $this->assertSame("pong", socket_read($socket, 4, PHP_BINARY_READ));
    }

    public function testTwoConnectionsReverseOrder()
    {
        $conn1 = $this->makeConnection($socket1);
        $this->assertTrue($conn1->isConnected());

        $conn2 = $this->makeConnection($socket2);
        $this->assertTrue($conn2->isConnected());

        socket_send($socket1, "ping", 4, 0);
        socket_send($socket2, "ping", 4, 0);

        $this->assertSame("pong", socket_read($socket2, 4, PHP_BINARY_READ));
        $this->assertSame("pong", socket_read($socket1, 4, PHP_BINARY_READ));
    }

    /**
     * @param resource $socket
     *
     * @return \Spiral\Goridge\Connection
     */
    protected function makeConnection(resource &$socket = null): Connection
    {
        $conn = new Connection(
            static::SOCK_ADDR,
            static::SOCK_PORT,
            static::SOCK_TYPE
        );

        $conn->connect();

        $connReflect = new \ReflectionObject($conn);
        $socketProp = $connReflect->getProperty("socket");

        $socketProp->setAccessible(true);
        $socket = $socketProp->getValue($conn);

        $this->assertInternalType("resource", $socket);

        return $conn;
    }
}