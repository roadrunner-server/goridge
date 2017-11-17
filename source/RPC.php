<?php
/**
 * Dead simple, high performance, drop-in bridge to Golang RPC with zero dependencies
 *
 * @author Wolfy-J
 */

namespace Spiral\Goridge;

use Spiral\Goridge\RelayInterface as Relay;

/**
 * RPC bridge to Golang net/rpc package over Goridge protocol.
 */
class RPC
{
    /** @var Relay */
    private $relay;

    /**
     * @param Relay $relay
     */
    public function __construct(Relay $relay)
    {
        $this->relay = $relay;
    }

    /**
     * @param string $method
     * @param mixed  $payload An binary data or array of arguments for complex types.
     * @param int    $flags   Payload control flags.
     *
     * @return mixed
     *
     * @throws Exceptions\RelayException
     * @throws Exceptions\ServiceException
     */
    public function call(string $method, $payload, int $flags = 0)
    {
        $this->relay->send($method, Relay::PAYLOAD_CONTROL | Relay::PAYLOAD_RAW);

        if ($flags & Relay::PAYLOAD_RAW) {
            $this->relay->send($payload, $flags);
        } else {
            $body = json_encode($payload);
            if ($body === false) {
                throw new Exceptions\ServiceException(sprintf("json encode: %s", json_last_error_msg()));
            }

            $this->relay->send($body);
        }

        // wait for the response
        $body = $this->relay->receiveSync($flags);

        return $this->handleBody($body, $flags);
    }

    /**
     * Handle response body.
     *
     * @param string $body
     * @param int    $flags
     *
     * @return mixed
     *
     * @throws Exceptions\ServiceException
     */
    protected function handleBody($body, int $flags)
    {
        if ($flags & Relay::PAYLOAD_ERROR && $flags & Relay::PAYLOAD_RAW) {
            throw new Exceptions\ServiceException("error '$body' on '{$this->relay}'");
        }

        if ($flags & Relay::PAYLOAD_RAW) {
            return $body;
        }

        return json_decode($body, true);
    }
}