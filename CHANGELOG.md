CHANGELOG
=========

## v2.0.3 (20.02.2018)
- fixed unix socket support on MacOS by @bgetsug

## v2.0.2 (29.01.2018)
- typo in SOCK_TPC constant

## v2.0.1 (23.01.2018)
- support sending empty string payloads without PAYLOAD_NONE flag

## v2.0.0 (17.11.2017)
- ext-sockets is not required anymore
- Connection renamed to Relay
- JsonRPC renamed to RPC
- Performance optimizations (+20% speed improvement)
- Added relay over standard streams and pipes
- Added ability to invoke PHP from Go
- Added control headers to manage execution flow
- CLOSE_CONNECTION and KEEP_CONNECTION constants are removed
- \*\_BODY constants renamed to PAYLOAD\_\* to comply with .go code 
- Protocol is extracted from Codec as independent abstraction
- Simplified RPC Codec implementation
- Code is formatted to comply to Golint
- More tests
- More documentation
- More error checks in PHP code
- License holder moved from Spiral Framework to SpiralScout

## v1.0.4 
- library dependency downgraded to support PHP7.0 by @thePanz

## v1.0.1 (14.08.2017)
- service level exception for invalid json payloads

## v1.0.0 (14.08.2017)
- public Goridge release
