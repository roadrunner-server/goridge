<?php

/**
 * Dead simple, high performance, drop-in bridge to Golang RPC with zero dependencies
 *
 * @author Wolfy-J
 */

declare(strict_types=1);

namespace Spiral\Goridge;

/**
 * Communicates with remote server/client over streams using byte payload:
 *
 * [ prefix       ][ payload                               ]
 * [ 1+8+8 bytes  ][ message length|LE ][message length|BE ]
 *
 * prefix:
 * [ flag       ][ message length, unsigned int 64bits, LittleEndian ]
 */
class StreamRelay implements RelayInterface, SendPackageRelayInterface
{
    /** @var resource */
    private $in;

    /** @var resource */
    private $out;

    /**
     * Example:
     * $relay = new StreamRelay(STDIN, STDOUT);
     *
     * @param resource $in  Must be readable.
     * @param resource $out Must be writable.
     *
     * @throws Exceptions\InvalidArgumentException
     */
    public function __construct($in, $out)
    {
        if (!is_resource($in) || get_resource_type($in) !== 'stream') {
            throw new Exceptions\InvalidArgumentException('expected a valid `in` stream resource');
        }

        if (!$this->assertReadable($in)) {
            throw new Exceptions\InvalidArgumentException('resource `in` must be readable');
        }

        if (!is_resource($out) || get_resource_type($out) !== 'stream') {
            throw new Exceptions\InvalidArgumentException('expected a valid `out` stream resource');
        }

        if (!$this->assertWritable($out)) {
            throw new Exceptions\InvalidArgumentException('resource `out` must be writable');
        }

        $this->in = $in;
        $this->out = $out;
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
        $headerPackage = packMessage($headerPayload, $headerFlags);
        $bodyPackage = packMessage($bodyPayload, $bodyFlags);
        if ($headerPackage === null || $bodyPackage === null) {
            throw new Exceptions\TransportException('unable to send payload with PAYLOAD_NONE flag');
        }

        if (
            fwrite(
                $this->out,
                $headerPackage['body'] . $bodyPackage['body'],
                34 + $headerPackage['size'] + $bodyPackage['size']
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
        $package = packMessage($payload, $flags);
        if ($package === null) {
            throw new Exceptions\TransportException('unable to send payload with PAYLOAD_NONE flag');
        }

        if (fwrite($this->out, $package['body'], 17 + $package['size']) === false) {
            throw new Exceptions\TransportException('unable to write payload to the stream');
        }

        return $this;
    }

    /**
     * {@inheritdoc}
     */
    public function receiveSync(?int &$flags = null): ?string
    {
        $prefix = $this->fetchPrefix();
        $flags = $prefix['flags'];

        $result = '';
        if ($prefix['size'] !== 0) {
            $leftBytes = $prefix['size'];

            //Add ability to write to stream in a future
            while ($leftBytes > 0) {
                $buffer = fread($this->in, min($leftBytes, self::BUFFER_SIZE));
                if ($buffer === false) {
                    throw new Exceptions\TransportException('error reading payload from the stream');
                }

                $result .= $buffer;
                $leftBytes -= strlen($buffer);
            }
        }

        return ($result !== '') ? $result : null;
    }

    /**
     * @return array Prefix [flag, length]
     *
     * @throws Exceptions\PrefixException
     */
    private function fetchPrefix(): array
    {
        $prefixBody = fread($this->in, 17);
        if ($prefixBody === false) {
            throw new Exceptions\PrefixException('unable to read prefix from the stream');
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
     * Checks if stream is readable.
     *
     * @param resource $stream
     *
     * @return bool
     */
    private function assertReadable($stream): bool
    {
        $meta = stream_get_meta_data($stream);

        return in_array($meta['mode'], ['r', 'rb', 'r+', 'rb+', 'w+', 'wb+', 'a+', 'ab+', 'x+', 'c+', 'cb+'], true);
    }

    /**
     * Checks if stream is writable.
     *
     * @param resource $stream
     *
     * @return bool
     */
    private function assertWritable($stream): bool
    {
        $meta = stream_get_meta_data($stream);

        return !in_array($meta['mode'], ['r', 'rb'], true);
    }
}
