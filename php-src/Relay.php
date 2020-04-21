<?php
/**
 * Dead simple, high performance, drop-in bridge to Golang RPC with zero dependencies
 *
 * @author Valentin V
 */

namespace Spiral\Goridge;

abstract class Relay
{
    public const TCP_SOCKET  = 'tcp';
    public const UNIX_SOCKET = 'unix';
    public const STREAM      = 'pipes';

    private const CONNECTION = '/(?P<protocol>[^:\/]+):\/\/(?P<arg1>[^:]+)(:(?P<arg2>[^:]+))?/';

    public static function create(string $connection): RelayInterface
    {
        if (!preg_match(self::CONNECTION, strtolower($connection), $match)) {
            throw new Exceptions\RelayFactoryException('unsupported connection format');
        }

        switch ($match['protocol']) {
            case self::TCP_SOCKET:
                //fall through
            case self::UNIX_SOCKET:
                return new SocketRelay(
                    $match['arg1'],
                    isset($match['arg2']) ? (int)$match['arg2'] : null,
                    $match['protocol'] === self::TCP_SOCKET ? SocketRelay::SOCK_TCP : SocketRelay::SOCK_UNIX
                );

            case self::STREAM:
                if (!isset($match['arg2'])) {
                    throw new Exceptions\RelayFactoryException('unsupported stream connection format');
                }

                return new StreamRelay(
                    fopen("php://{$match['arg1']}", 'rb'),
                    fopen("php://{$match['arg2']}", 'wb')
                );
            default:
                throw new Exceptions\RelayFactoryException('unknown connection protocol');
        }
    }
}
