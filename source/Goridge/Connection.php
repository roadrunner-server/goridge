<?php
/**
 * Dead simple, high performance, drop-in bridge to Golang RPC with zero dependencies
 *
 * @author Wolfy-J
 */

namespace Spiral\Goridge;

use Spiral\Goridge\Exceptions\InvalidArgumentException;
use Spiral\Goridge\Exceptions\MessageException;
use Spiral\Goridge\Exceptions\PrefixException;
use Spiral\Goridge\Exceptions\TransportException;

/**
 * Sends byte payload to RCP server using protocol:
 *
 * [ prefix     ][ payload        ]
 * [ 1+8 bytes  ][ message length ]
 *
 * prefix:
 * [ flag       ][ message length, unsigned int 64bits, LittleEndian ]
 *
 * flag options:
 * KEEP_CONNECTION  - keep socket connection.
 * CLOSE_CONNECTION - end socket connection.
 */
class Connection implements ConnectionInterface
{
    const CHUNK_SIZE = 65536;

    /** Supported socket types. */
    const SOCK_TPC  = 0;
    const SOCK_UNIX = 1;

    /** @var string */
    private $address;

    /** @var int|null */
    private $port;

    /** @var int */
    private $type;

    /** @var resource|null */
    private $socket;

    /**
     * Example:
     * $conn = new Connection("localhost", 7000);
     *
     * $conn = new Connection("/tmp/rpc.sock", null, Socket::UNIX_SOCKET);
     *
     * @param string   $address Localhost, ip address or hostname.
     * @param int|null $port    Ignored for UNIX sockets.
     * @param int      $type    Default: TPC_SOCKET
     *
     * @throws \Spiral\Goridge\Exceptions\InvalidArgumentException
     */
    public function __construct(string $address, int $port = null, int $type = self::SOCK_TPC)
    {
        switch ($type) {
            case self::SOCK_TPC:
                if ($port === null) {
                    throw new InvalidArgumentException(sprintf(
                        "no port given for TPC socket on '%s'",
                        $address
                    ));
                }
                break;
            case self::SOCK_UNIX:
                $port = null;
                break;
            default:
                throw new InvalidArgumentException(sprintf(
                    "undefined connection type %s on '%s'",
                    $type,
                    $address
                ));
        }

        $this->address = $address;
        $this->port = $port;
        $this->type = $type;
    }

    /**
     * @return string
     */
    public function getAddress(): string
    {
        return $this->address;
    }

    /**
     * @return int|null
     */
    public function getPort(): ? int
    {
        return $this->port;
    }

    /**
     * @return int
     */
    public function getType(): int
    {
        return $this->type;
    }

    /**
     * @return bool
     */
    public function isConnected(): bool
    {
        return $this->socket != null;
    }

    /**
     * Send payload message to another party.
     *
     * @param string|binary $payload
     * @param int           $flags
     *
     * @return self
     *
     * @throws \Spiral\Goridge\Exceptions\MessageException When message can not be send.
     */
    public function send($payload, int $flags = self::KEEP_CONNECTION): self
    {
        $this->connect();

        $size = strlen($payload);

        if ($flags & self::NO_BODY && $size != 0) {
            throw new MessageException("unable to set body with NO_BODY flag");
        }

        socket_send($this->socket, pack('CP', $flags, $size), 10, 0);

        if (!($flags & self::NO_BODY)) {
            socket_send($this->socket, $payload, $size, 0);
        }

        return $this;
    }

    /**
     * Receive message from another party in sync/blocked mode. Message can be null.
     *
     * @param int $flags Response flags.
     *
     * @return null|string
     *
     * @throws \Spiral\Goridge\Exceptions\TransportException When unable to connect or maintain
     *                                                         socket.
     * @throws \Spiral\Goridge\Exceptions\MessageException When messages can not be retrieved.
     */
    public function receiveSync(int & $flags = null): ? string
    {
        $this->connect();

        $prefix = $this->fetchPrefixSync();
        $flags = $prefix['flags'];
        $result = null;

        if ($prefix['size'] !== 0) {
            $readBytes = $prefix['size'];
            $buffer = null;

            //Add ability to write to stream in a future
            while ($readBytes > 0) {
                $bufferLength = socket_recv(
                    $this->socket,
                    $buffer,
                    min(self::CHUNK_SIZE, $readBytes),
                    MSG_WAITALL
                );

                $result .= $buffer;
                $readBytes -= $bufferLength;
            }
        }

        if ($flags & self::CLOSE_CONNECTION) {
            $this->close();
        }

        return $result;
    }

    /**
     * Ensure socket connection. Returns true if socket successfully connected
     * or have already been connected.
     *
     * @return bool
     *
     * @throws \Spiral\Goridge\Exceptions\TransportException
     * @throws \Error When sockets are used in unsupported environment.
     */
    public function connect(): bool
    {
        if ($this->isConnected()) {
            return true;
        }

        $this->socket = $this->createSocket();
        try {
            if (socket_connect($this->socket, $this->address, $this->port) === false) {
                throw new TransportException(socket_strerror(socket_last_error($this->socket)));
            }
        } catch (\Exception $e) {
            throw new TransportException("unable to establish connection {$this}", 0, $e);
        }

        return true;
    }

    /**
     * Close connection.
     *
     * @throws \Spiral\Goridge\Exceptions\TransportException
     */
    public function close()
    {
        if (!$this->isConnected()) {
            throw new TransportException("unable to close socket '{$this}', socket already closed");
        }

        socket_close($this->socket);
        $this->socket = null;
    }

    /**
     * Destruct connection and disconnect.
     */
    public function __destruct()
    {
        if ($this->isConnected()) {
            $this->close();
        }
    }

    /**
     * @return string
     */
    public function __toString(): string
    {
        if ($this->type == self::SOCK_TPC) {
            return "tcp://{$this->address}:{$this->port}";
        }

        return "unix://{$this->address}";
    }

    /**
     * @return array Prefix [flag, length]
     *
     * @throws \Spiral\Goridge\Exceptions\PrefixException
     */
    private function fetchPrefixSync(): array
    {
        $prefixLength = socket_recv($this->socket, $prefixBody, 9, MSG_WAITALL);
        if ($prefixBody === false || $prefixLength !== 9) {
            throw new PrefixException(sprintf(
                "unable to read prefix from socket: %s",
                socket_strerror(socket_last_error($this->socket))
            ));
        }

        return unpack("Cflags/Psize", $prefixBody);
    }

    /**
     * @return resource
     * @throws \Error
     */
    private function createSocket()
    {
        if ($this->type === self::SOCK_UNIX) {
            if (strtoupper(substr(PHP_OS, 0, 3)) === 'WIN') {
                throw new \Error("socket {$this} unavailable in windows");
            }

            return socket_create(1, SOCK_STREAM, SOL_SOCKET);

        }

        return socket_create(AF_INET, SOCK_STREAM, SOL_TCP);
    }
}