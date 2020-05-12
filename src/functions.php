<?php

/**
 * Dead simple, high performance, drop-in bridge to Golang RPC with zero dependencies
 *
 * @author Valentin V
 */

declare(strict_types=1);

namespace Spiral\Goridge;

if (!function_exists('Spiral\\Goridge\\packMessage')) {
    /**
     * @param string   $payload
     * @param int|null $flags
     * @return array|null
     * @internal
     */
    function packMessage(string $payload, ?int $flags = null): ?array
    {
        $size = strlen($payload);
        if ($flags & RelayInterface::PAYLOAD_NONE && $size !== 0) {
            return null;
        }

        $body = pack('CPJ', $flags, $size, $size);

        if (!($flags & RelayInterface::PAYLOAD_NONE)) {
            $body .= $payload;
        }

        return compact('body', 'size');
    }
}
