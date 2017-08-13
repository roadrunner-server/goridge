<?php
/**
 * Dead simple, high performance, drop-in bridge to Golang RPC with zero dependencies
 *
 * @author Wolfy-J
 */

namespace Spiral\Goridge;

use Spiral\Goridge\Exceptions\ServiceException;

/**
 * Packs client requests and responses using JSON.
 */
class JsonRPC
{
    /** @var \Spiral\Goridge\ConnectionInterface */
    private $conn;

    /**
     * @param \Spiral\Goridge\ConnectionInterface $conn
     */
    public function __construct(ConnectionInterface $conn)
    {
        $this->conn = $conn;
    }

    /**
     * @param string $method
     * @param mixed  $argument An input argument or array of arguments for complex types.
     * @param int    $flags    Payload control flags.
     *
     * @return mixed
     *
     * @throws \Spiral\Goridge\Exceptions\TransportException
     * @throws \Spiral\Goridge\Exceptions\ServiceException
     */
    public function call(string $method, $argument, int $flags = 0)
    {
        $this->conn->send($method);

        if ($flags & ConnectionInterface::RAW_BODY) {
            $this->conn->send($argument, $flags);
        } else {
            $body = json_encode($argument);
            if ($body === false) {
                throw new ServiceException(sprintf(
                    "json encode: %s",
                    json_last_error_msg()
                ));
            }
            
            $this->conn->send(json_encode($argument));
        }

        $body = $this->conn->receiveSync($flags);

        return $this->handleBody($body, $flags);
    }

    /**
     * Handle response body.
     *
     * @param string|binary $body
     * @param int           $flags
     *
     * @return mixed
     *
     * @throws \Spiral\Goridge\Exceptions\ServiceException
     */
    protected function handleBody($body, int $flags)
    {
        if ($flags & ConnectionInterface::ERROR_BODY) {
            throw new ServiceException("error '$body' on '{$this->conn}'");
        }

        if ($flags & ConnectionInterface::RAW_BODY) {
            return $body;
        }

        return json_decode($body, true);
    }
}
