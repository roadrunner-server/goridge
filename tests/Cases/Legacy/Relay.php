<?php

namespace Spiral\Tests\Legacy;

use Spiral\Goridge\RelayInterface;

class Relay implements RelayInterface
{
    /** @var RelayInterface */
    private $relay;

    public function __construct(RelayInterface $relay)
    {
        $this->relay = $relay;
    }

    /**
     * @return string
     */
    public function __toString(): string
    {
        return (string)$this->relay;
    }

    public function send($payload, int $flags = null)
    {
        return $this->relay->send($payload, $flags);
    }

    public function receiveSync(int &$flags = null)
    {
        return $this->relay->receiveSync($flags);
    }
}
