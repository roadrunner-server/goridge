<?php

/**
 * Dead simple, high performance, drop-in bridge to Golang RPC with zero dependencies
 *
 * @author Valentin V
 */

declare(strict_types=1);

namespace Spiral\Tests;

use PHPUnit\Framework\TestCase;
use Spiral\Goridge\Exceptions;
use Spiral\Goridge\Relay;
use Spiral\Goridge\SocketRelay;
use Spiral\Goridge\StreamRelay;
use Throwable;

class RelayFactoryTest extends TestCase
{
    /**
     * @dataProvider formatProvider
     * @param string $connection
     * @param bool   $expectedException
     */
    public function testFormat(string $connection, bool $expectedException = false): void
    {
        $this->assertTrue(true);
        if ($expectedException) {
            $this->expectException(Exceptions\RelayFactoryException::class);
        }

        try {
            Relay::create($connection);
        } catch (Exceptions\RelayFactoryException $exception) {
            throw $exception;
        } catch (Throwable $exception) {
            //do nothing, that's not a factory issue
        }
    }

    /**
     * @return iterable
     */
    public function formatProvider(): iterable
    {
        return [
            //format invalid
            ['tcp:localhost:', true],
            ['tcp:/localhost:', true],
            ['tcp//localhost:', true],
            ['tcp//localhost', true],
            //unknown provider
            ['test://localhost', true],
            //pipes require 2 args
            ['pipes://localhost:', true],
            ['pipes://localhost', true],
            //valid format
            ['tcp://localhost'],
            ['tcp://localhost:123'],
            ['unix://localhost:123'],
            ['unix://rpc.sock'],
            ['unix:///tmp/rpc.sock'],
            ['tcp://localhost:abc'],
            ['pipes://stdin:stdout'],
        ];
    }

    public function testTCP(): void
    {
        /** @var SocketRelay $relay */
        $relay = Relay::create('tcp://localhost:0');
        $this->assertInstanceOf(SocketRelay::class, $relay);
        $this->assertSame('localhost', $relay->getAddress());
        $this->assertSame(0, $relay->getPort());
        $this->assertSame(SocketRelay::SOCK_TCP, $relay->getType());
    }

    public function testUnix(): void
    {
        /** @var SocketRelay $relay */
        $relay = Relay::create('unix:///tmp/rpc.sock');
        $this->assertInstanceOf(SocketRelay::class, $relay);
        $this->assertSame('/tmp/rpc.sock', $relay->getAddress());
        $this->assertSame(SocketRelay::SOCK_UNIX, $relay->getType());
    }

    public function testPipes(): void
    {
        /** @var StreamRelay $relay */
        $relay = Relay::create('pipes://stdin:stdout');
        $this->assertInstanceOf(StreamRelay::class, $relay);
    }
}
