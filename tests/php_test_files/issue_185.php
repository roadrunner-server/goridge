<?php

use Spiral\Goridge;
require "vendor/autoload.php";

$rpc = new Goridge\RPC\RPC(
    Goridge\Relay::create('tcp://127.0.0.1:6001')
);

//or, using factory:
$tcpRPC = new Goridge\RPC\RPC(Goridge\Relay::create('tcp://127.0.0.1:6001'));
$unixRPC = new Goridge\RPC\RPC(Goridge\Relay::create('unix:///tmp/rpc.sock'));
$streamRPC = new Goridge\RPC\RPC(Goridge\Relay::create('pipes://stdin:stdout'));

echo $rpc->call("App.Hi", "Antony");