<?php

/**
 * Dead simple, high performance, drop-in bridge to Golang RPC with zero dependencies
 *
 * @author Wolfy-J
 */

declare(strict_types=1);

namespace Spiral\Tests;

use Exception;
use PHPUnit\Framework\TestCase;
use Spiral\Goridge\Exceptions\ServiceException;
use Spiral\Goridge\RelayInterface;
use Spiral\Goridge\RPC;
use Spiral\Goridge\SocketRelay;

abstract class RPCTest extends TestCase
{
    public const GO_APP    = 'server';
    public const SOCK_ADDR = '';
    public const SOCK_PORT = 7079;
    public const SOCK_TYPE = SocketRelay::SOCK_TCP;

    public function testPingPong(): void
    {
        $conn = $this->makeRPC();
        $this->assertSame('pong', $conn->call('Service.Ping', 'ping'));
    }

    public function testPingNull(): void
    {
        $conn = $this->makeRPC();
        $this->assertSame('', $conn->call('Service.Ping', 'not-ping'));
    }

    public function testNegate(): void
    {
        $conn = $this->makeRPC();
        $this->assertSame(-10, $conn->call('Service.Negate', 10));
    }

    public function testNegateNegative(): void
    {
        $conn = $this->makeRPC();
        $this->assertSame(10, $conn->call('Service.Negate', -10));
    }

    /**
     * @throws Exception
     */
    public function testLongEcho(): void
    {
        $conn = $this->makeRPC();
        $payload = base64_encode(random_bytes(SocketRelay::BUFFER_SIZE * 5));

        $resp = $conn->call('Service.Echo', $payload);

        $this->assertSame(strlen($payload), strlen($resp));
        $this->assertSame(md5($payload), md5($resp));
    }

    /**
     * @throws Exception
     */
    public function testConvertException(): void
    {
        $this->expectException(ServiceException::class);
        $this->expectExceptionMessage('{rawData} request for <*string Value>');

        $conn = $this->makeRPC();
        $payload = base64_encode(random_bytes(SocketRelay::BUFFER_SIZE * 5));

        $resp = $conn->call(
            'Service.Echo',
            $payload,
            RelayInterface::PAYLOAD_RAW
        );

        $this->assertSame(strlen($payload), strlen($resp));
        $this->assertSame(md5($payload), md5($resp));
    }

    /**
     * @throws Exception
     */
    public function testRawBody(): void
    {
        $conn = $this->makeRPC();
        $payload = random_bytes(SocketRelay::BUFFER_SIZE * 100);

        $resp = $conn->call(
            'Service.EchoBinary',
            $payload,
            RelayInterface::PAYLOAD_RAW
        );

        $this->assertSame(strlen($payload), strlen($resp));
        $this->assertSame(md5($payload), md5($resp));
    }

    /**
     * @throws Exception
     */
    public function testLongRawBody(): void
    {
        $conn = $this->makeRPC();
        $payload = random_bytes(SocketRelay::BUFFER_SIZE * 100);

        $resp = $conn->call(
            'Service.EchoBinary',
            $payload,
            RelayInterface::PAYLOAD_RAW
        );

        $this->assertSame(strlen($payload), strlen($resp));
        $this->assertSame(md5($payload), md5($resp));
    }

    public function testPayload(): void
    {
        $conn = $this->makeRPC();

        $resp = $conn->call('Service.Process', [
            'name'  => 'wolfy-j',
            'value' => 18
        ]);

        $this->assertSame([
            'name'  => 'WOLFY-J',
            'value' => -18
        ], $resp);
    }

    public function testBadPayload(): void
    {
        $this->expectException(ServiceException::class);
        $this->expectExceptionMessage('{rawData} request for <*main.Payload Value>');

        $conn = $this->makeRPC();
        $conn->call('Service.Process', 'raw', RelayInterface::PAYLOAD_RAW);
    }

    public function testPayloadWithMap(): void
    {
        $conn = $this->makeRPC();

        $resp = $conn->call('Service.Process', [
            'name'  => 'wolfy-j',
            'value' => 18,
            'keys'  => [
                'key'   => 'value',
                'email' => 'domain'
            ]
        ]);

        $this->assertIsArray($resp['keys']);
        $this->assertArrayHasKey('value', $resp['keys']);
        $this->assertArrayHasKey('domain', $resp['keys']);

        $this->assertSame('key', $resp['keys']['value']);
        $this->assertSame('email', $resp['keys']['domain']);
    }

    public function testBrokenPayloadMap(): void
    {
        $this->expectException(ServiceException::class);

        $conn = $this->makeRPC();

        $conn->call('Service.Process', [
            'name'  => 'wolfy-j',
            'value' => 18,
            'keys'  => 1111
        ]);
    }

    /**
     * @throws Exception
     */
    public function testJsonException(): void
    {
        $this->expectException(ServiceException::class);
        $this->expectExceptionMessageMatches('#.*json encode.*#');

        $conn = $this->makeRPC();

        $conn->call('Service.Process', random_bytes(256));
    }

    /**
     * @return RPC
     */
    protected function makeRPC(): RPC
    {
        return new RPC(new SocketRelay(static::SOCK_ADDR, static::SOCK_PORT, static::SOCK_TYPE));
    }
}
