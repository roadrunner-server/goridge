<?php
/**
 * Dead simple, high performance, drop-in bridge to Golang RPC with zero dependencies
 *
 * @author Wolfy-J
 */

namespace Spiral\Goridge;

use Spiral\Goridge\Exceptions\MessageException;
use Spiral\Goridge\Exceptions\TransportException;

interface ConnectionInterface
{
    /** Message delivery flags. */
    const KEEP_CONNECTION  = 0;
    const CLOSE_CONNECTION = 1;

    /** Payload flasg.*/
    const NO_BODY    = 16;
    const RAW_BODY   = 32;
    const ERROR_BODY = 64;

    /**
     * Send payload message to another party.
     *
     * @param string|binary $payload
     * @param int           $flags Protocol control flags.
     *
     * @throws MessageException When message can not be send.
     * @throws TransportException
     */
    public function send($payload, int $flags = self::KEEP_CONNECTION);

    /**
     * Receive message from another party in sync/blocked mode. Message can be null.
     *
     * @param int $flags Response flags.
     *
     * @return null|string
     *
     * @throws MessageException When messages can not be retrieved.
     * @throws TransportException
     */
    public function receiveSync(int &$flags = null);
}