<?php
use Spiral\Goridge;

require "vendor/autoload.php";

$rpc = new Goridge\RPC(
  new Goridge\SocketRelay("127.0.0.1", 6001)
);

echo $rpc->call("App.Hi", "Antony");
