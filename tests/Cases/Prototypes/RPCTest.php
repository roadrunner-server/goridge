<?php
/**
 * goridge
 *
 * @author    Wolfy-J
 */

namespace Spiral\Tests\Prototypes;

use Spiral\Goridge\Connection;
use Spiral\Goridge\ConnectionInterface;
use Spiral\Goridge\JsonRPC;

abstract class RPCTest extends ProcessTest
{
    const GO_APP    = "server";
    const SOCK_ADDR = "";
    const SOCK_PORT = 7079;
    const SOCK_TYPE = Connection::SOCK_TPC;

    public function testPingPong()
    {
        $conn = $this->makeRPC();
        $this->assertSame("pong", $conn->call('Service.Ping', 'ping'));
    }

    public function testPingNull()
    {
        $conn = $this->makeRPC();
        $this->assertSame("", $conn->call('Service.Ping', 'not-ping'));
    }

    public function testNegate()
    {
        $conn = $this->makeRPC();
        $this->assertSame(-10, $conn->call('Service.Negate', 10));
    }

    public function testNegateNegative()
    {
        $conn = $this->makeRPC();
        $this->assertSame(10, $conn->call('Service.Negate', -10));
    }

    public function testLongEcho()
    {
        $conn = $this->makeRPC();
        $payload = base64_encode(random_bytes(Connection::CHUNK_SIZE * 5));

        $resp = $conn->call('Service.Echo', $payload);

        $this->assertSame(strlen($payload), strlen($resp));
        $this->assertSame(md5($payload), md5($resp));
    }

    /**
     * @expectedException \Spiral\Goridge\Exceptions\ServiceException
     * @expectedExceptionMessageRegExp #error '{rawData} request for string'.*#
     */
    public function testConvertException()
    {
        $conn = $this->makeRPC();
        $payload = base64_encode(random_bytes(Connection::CHUNK_SIZE * 5));

        $resp = $conn->call(
            'Service.Echo',
            $payload,
            ConnectionInterface::RAW_BODY
        );

        $this->assertSame(strlen($payload), strlen($resp));
        $this->assertSame(md5($payload), md5($resp));
    }

    public function testRawBody()
    {
        $conn = $this->makeRPC();
        $payload = random_bytes(Connection::CHUNK_SIZE * 100);

        $resp = $conn->call(
            'Service.EchoBinary',
            $payload,
            ConnectionInterface::RAW_BODY
        );

        $this->assertSame(strlen($payload), strlen($resp));
        $this->assertSame(md5($payload), md5($resp));
    }

    public function testLongRawBody()
    {
        $conn = $this->makeRPC();
        $payload = random_bytes(Connection::CHUNK_SIZE * 1000);

        $resp = $conn->call(
            'Service.EchoBinary',
            $payload,
            ConnectionInterface::RAW_BODY
        );

        $this->assertSame(strlen($payload), strlen($resp));
        $this->assertSame(md5($payload), md5($resp));
    }

    public function testPayload()
    {
        $conn = $this->makeRPC();

        $resp = $conn->call('Service.Process', [
            'name'  => "wolfy-j",
            'value' => 18
        ]);

        $this->assertSame([
            'name'  => "WOLFY-J",
            'value' => -18
        ], $resp);
    }

    /**
     * @expectedException \Spiral\Goridge\Exceptions\ServiceException
     * @expectedExceptionMessageRegExp #error '{rawData} request for struct.*#
     */
    public function testBadPayload()
    {
        $conn = $this->makeRPC();
        $conn->call('Service.Process', 'raw', ConnectionInterface::RAW_BODY);
    }

    public function testPayloadWithMap()
    {
        $conn = $this->makeRPC();

        $resp = $conn->call('Service.Process', [
            'name'  => "wolfy-j",
            'value' => 18,
            'keys'  => [
                'key'   => 'value',
                'email' => 'domain'
            ]
        ]);

        $this->assertInternalType('array', $resp['keys']);
        $this->assertArrayHasKey('value', $resp['keys']);
        $this->assertArrayHasKey('domain', $resp['keys']);

        $this->assertSame('key', $resp['keys']['value']);
        $this->assertSame('email', $resp['keys']['domain']);
    }

    /**
     * @expectedException \Spiral\Goridge\Exceptions\ServiceException
     * @expectedExceptionMessageRegExp #error '.*cannot unmarshal number into Go struct field Payload.keys.*#
     */
    public function testBrokenPayloadMap()
    {
        $conn = $this->makeRPC();

        $resp = $conn->call('Service.Process', [
            'name'  => "wolfy-j",
            'value' => 18,
            'keys'  => 1111
        ]);
    }

    /**
     * @return \Spiral\Goridge\JsonRPC
     */
    protected function makeRPC(): JsonRPC
    {
        return new JsonRPC(new Connection(
            static::SOCK_ADDR,
            static::SOCK_PORT,
            static::SOCK_TYPE
        ));
    }
}