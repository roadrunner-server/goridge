CHANGELOG
=========

## v2.3.0 (23.03.2020)
- Replaced std encoding/json package with the https://github.com/json-iterator/go 
- Added BORS and GHA support
- golang modules updated to v2

## v2.2.1 (30.11.2019)
- Fixed too strict StreamRelay check by @tarampampam 

## v2.2.0 (29.11.2019)
- Updated travis to support go 1.13
- Updated tests (errors handling, simplify)
- Add go modules support
- Optimize pack function

by @ValeryPiashchynski 

## v2.1.4 (01.04.2019)
- minor performance improvements by @zloyuser
- removed nighly from travis

## v2.1.3 (30.09.2018)
- improved performance (reduced number of syscalls required for Send command)

## v2.1.2 (07.06.2018)
- added 8 more bytes to the payload
- added error detection mechanism over binary masks
- added panic handler for pipe relay

## v2.1.0 (03.06.2018)
- added golang ClientCodec implementation
- additional error detections
- added sequence support
- more tests

## v2.0.5 (03.04.2018)
- handled possible panic on reading from broken connection in socket relay

## v2.0.4 (23.03.2018)
- minor performance improvement in memory allocation for buffer by @243083df

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
