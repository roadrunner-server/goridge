<?php

use Spiral\Goridge;

require "../vendor/autoload.php";

$rpc = new Goridge\JsonRPC(new Goridge\Connection("127.0.0.1", 6001));

echo $rpc->call("App.Hi", "Antony");