<?php
/**
 * Dead simple, high performance, drop-in bridge to Golang RPC with zero dependencies
 *
 * @author Valentin V
 */

namespace Spiral\Goridge;

abstract class Relay implements RelayInterface
{
    public const TCP_SOCKET  = 'tcp';
    public const UNIX_SOCKET = 'unix';
    public const STREAM      = 'pipes';

    public static function create(string $connection): RelayInterface
    {
        $parsed = self::detectProtocol($connection);
        if ($parsed === null) {
            throw new Exceptions\RelayException('unknown connection');
        }

        [$protocol, $etc] = $parsed;

    }

    private static function detectProtocol(string $connection)
    {
        $connection = strtolower($connection);
        if (mb_strpos($connection, '://') === false) {
            return null;
        }

        return explode('://', $connection, 2);
    }

    private static function createTCP(string $connection): SocketRelay
    {
    }
    private static function createUnix(string $connection): SocketRelay
    {
    }
}
