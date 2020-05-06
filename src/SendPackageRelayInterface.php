<?php

/**
 * Dead simple, high performance, drop-in bridge to Golang RPC with zero dependencies
 *
 * @author Wolfy-J
 */

declare(strict_types=1);

namespace Spiral\Goridge;

interface SendPackageRelayInterface
{
    /**
     * Send message package with header and body.
     *
     * @param string   $headerPayload
     * @param int|null $headerFlags
     * @param string   $bodyPayload
     * @param int|null $bodyFlags
     * @return mixed
     */
    public function sendPackage(
        string $headerPayload,
        ?int $headerFlags,
        string $bodyPayload,
        ?int $bodyFlags = null
    );
}
