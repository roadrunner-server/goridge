<?php

/**
 * Dead simple, high performance, drop-in bridge to Golang RPC with zero dependencies
 *
 * @author Wolfy-J
 */

declare(strict_types=1);

namespace Spiral\Goridge;

use Error;
use Exception;

/**
 * Communicates with remote server/client over be-directional socket using byte payload:
 *
 * [ prefix       ][ payload                               ]
 * [ 1+8+8 bytes  ][ message length|LE ][message length|BE ]
 *
 * prefix:
 * [ flag       ][ message length, unsigned int 64bits, LittleEndian ]
 */
class SocketRelay implements RelayInterface, SendPackageRelayInterface, StringableRelayInterface
{
    /** Supported socket types. */
    public const SOCK_TCP  = 0;
    public const SOCK_UNIX = 1;

    // @deprecated
    public const SOCK_TPC = self::SOCK_TCP;

    /** @var string */
    private $address;

    /** @var int|null */
    private $port;

    /** @var int */
    private $type;

    /** @var resource */
    private $socket;

    /** @var bool */
    private $connected = false;

    /**
     * Example:
     * $relay = new SocketRelay("localhost", 7000);
     * $relay = new SocketRelay("/tmp/rpc.sock", null, Socket::UNIX_SOCKET);
     *
     * @param string   $address Localhost, ip address or hostname.
     * @param int|null $port    Ignored for UNIX sockets.
     * @param int      $type    Default: TCP_SOCKET
     *
     * @throws Exceptions\InvalidArgumentException
     */
    public function __construct(string $address, ?int $port = null, int $type = self::SOCK_TCP)
    {
        if (!extension_loaded('sockets')) {
            throw new Exceptions\InvalidArgumentException("'sockets' extension not loaded");
        }

        switch ($type) {
            case self::SOCK_TCP:
                if ($port === null) {
                    throw new Exceptions\InvalidArgumentException(sprintf(
                        "no port given for TPC socket on '%s'",
                        $address
                    ));
                }

                if ($port < 0 || $port > 65535) {
                    throw new Exceptions\InvalidArgumentException(sprintf(
                        "invalid port given for TPC socket on '%s'",
                        $address
                    ));
                }
                break;
            case self::SOCK_UNIX:
                $port = null;
                break;
            default:
                throw new Exceptions\InvalidArgumentException(sprintf(
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
        if ($this->type === self::SOCK_TCP) {
            return "tcp://{$this->address}:{$this->port}";
        }

        return "unix://{$this->address}";
    }

    /**
     * Send message package with header and body.
     *
     * @param string   $headerPayload
     * @param int|null $headerFlags
     * @param string   $bodyPayload
     * @param int|null $bodyFlags
     * @return self
     */
    public function sendPackage(
        string $headerPayload,
        ?int $headerFlags,
        string $bodyPayload,
        ?int $bodyFlags = null
    ): self {
        $this->connect();

        $headerPackage = packMessage($headerPayload, $headerFlags);
        $bodyPackage = packMessage($bodyPayload, $bodyFlags);
        if ($headerPackage === null || $bodyPackage === null) {
            throw new Exceptions\TransportException('unable to send payload with PAYLOAD_NONE flag');
        }

        if (
            socket_send(
                $this->socket,
                $headerPackage['body'] . $bodyPackage['body'],
                34 + $headerPackage['size'] + $bodyPackage['size'],
                0
            ) === false
        ) {
            throw new Exceptions\TransportException('unable to write payload to the stream');
        }

        return $this;
    }

    /**
     * {@inheritdoc}
     * @return self
     */
    public function send(string $payload, ?int $flags = null): self
    {
        $this->connect();

        $package = packMessage($payload, $flags);
        if ($package === null) {
            throw new Exceptions\TransportException('unable to send payload with PAYLOAD_NONE flag');
        }

        if (socket_send($this->socket, $package['body'], 17 + $package['size'], 0) === false) {
            throw new Exceptions\TransportException('unable to write payload to the stream');
        }

        return $this;
    }

    /**
     * {@inheritdoc}
     */
    public function receiveSync(?int &$flags = null): ?string
    {
        $this->connect();

        $prefix = $this->fetchPrefix();
        $flags = $prefix['flags'];

        $result = '';
        if ($prefix['size'] !== 0) {
            $readBytes = $prefix['size'];

            //Add ability to write to stream in a future
            while ($readBytes > 0) {
                $bufferLength = socket_recv(
                    $this->socket,
                    $buffer,
                    min(self::BUFFER_SIZE, $readBytes),
                    MSG_WAITALL
                );
                if ($bufferLength === false || $buffer === null) {
                    throw new Exceptions\PrefixException(sprintf(
                        'unable to read prefix from socket: %s',
                        socket_strerror(socket_last_error($this->socket))
                    ));
                }

                $result .= $buffer;
                $readBytes -= $bufferLength;
            }
        }

        return ($result !== '') ? $result : null;
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
    public function getPort(): ?int
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
        return $this->connected;
    }

    /**
     * Ensure socket connection. Returns true if socket successfully connected
     * or have already been connected.
     *
     * @return bool
     *
     * @throws Exceptions\RelayException
     * @throws Error When sockets are used in unsupported environment.
     */
    public function connect(): bool
    {
        if ($this->isConnected()) {
            return true;
        }

        $socket = $this->createSocket();
        if ($socket === false) {
            throw new Exceptions\RelayException("unable to create socket {$this}");
        }

        try {
            if (socket_connect($socket, $this->address, $this->port ?? 0) === false) {
                throw new Exceptions\RelayException(socket_strerror(socket_last_error($socket)));
            }
        } catch (Exception $e) {
            throw new Exceptions\RelayException("unable to establish connection {$this}", 0, $e);
        }

        $this->socket = $socket;
        $this->connected = true;

        return true;
    }

    /**
     * Close connection.
     *
     * @throws Exceptions\RelayException
     */
    public function close(): void
    {
        if (!$this->isConnected()) {
            throw new Exceptions\RelayException("unable to close socket '{$this}', socket already closed");
        }

        socket_close($this->socket);
        $this->connected = false;
        unset($this->socket);
    }

    /**
     * @return array Prefix [flag, length]
     *
     * @throws Exceptions\PrefixException
     */
    private function fetchPrefix(): array
    {
        $prefixLength = socket_recv($this->socket, $prefixBody, 17, MSG_WAITALL);
        if ($prefixBody === null || $prefixLength !== 17) {
            throw new Exceptions\PrefixException(sprintf(
                'unable to read prefix from socket: %s',
                socket_strerror(socket_last_error($this->socket))
            ));
        }

        $result = unpack('Cflags/Psize/Jrevs', $prefixBody);
        if (!is_array($result)) {
            throw new Exceptions\PrefixException('invalid prefix');
        }

        if ($result['size'] !== $result['revs']) {
            throw new Exceptions\PrefixException('invalid prefix (checksum)');
        }

        return $result;
    }

    /**
     * @return resource|false
     * @throws Exceptions\GoridgeException
     */
    private function createSocket()
    {
        if ($this->type === self::SOCK_UNIX) {
            return socket_create(AF_UNIX, SOCK_STREAM, 0);
        }

        return socket_create(AF_INET, SOCK_STREAM, SOL_TCP);
    }
}
