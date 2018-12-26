<?php
/**
 * Dead simple, high performance, drop-in bridge to Golang RPC with zero dependencies
 *
 * @author Wolfy-J
 */

namespace Spiral\Goridge;

use Spiral\Goridge\Exceptions\RelayException;
use Spiral\Goridge\Exceptions\TransportException;

interface RelayInterface
{
    /** Maximum payload size to read at once. */
    const BUFFER_SIZE = 65536;

    /** Payload flags.*/
    const PAYLOAD_NONE = 2;
    const PAYLOAD_RAW = 4;
    const PAYLOAD_ERROR = 8;
    const PAYLOAD_CONTROL = 16;

    /**
     * Send payload message to another party.
     *
     * @param string   $payload
     * @param int|null $flags Protocol control flags.
     *
     * @throws TransportException
     */
    public function send($payload, int $flags = null);

    /**
     * Receive message from another party in sync/blocked mode. Message can be null.
     *
     * @param int|null $flags Response flags.
     *
     * @return null|string
     *
     * @throws TransportException
     */
    public function receiveSync(int &$flags = null);
}
