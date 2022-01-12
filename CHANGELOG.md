CHANGELOG
=========

v3.2.5 (12.01.2022)
-------------------

## 📦 Packages:

- ✏️ Replace `jsoniter` with `goccy/json`

---

v3.2.4 (25.11.2021)
-------------------

## 🚀 New:

- ✏️ Include bad header data in the returned error.

---

v3.2.3 (18.10.2021)
-------------------

## 🚀 New:

- ✏️ New reusable pool with size classes. More than 50% boost in allocation time.
- ✏️ Better handling `file not found` error.

---

v3.2.2 (14.10.2021)
-------------------

## 🚀 New:

- 🚀 Internal performance improvements (~15% for the RPC).

---

v3.2.1 (19.08.2021)
-------------------

## 📦 Packages:

- 📦 Update Go to `v1.17`
- 📦 Update `msgpack` to `v5`

---

v3.2.0 (09.08.2021)
-------------------

## 🚀 New:

- ✏️ 50% reduce bound checks in the frame's operations. [PR](https://github.com/spiral/goridge/pull/143)
-

## 🩹 Fixes:

- 🐛 Fix: bug with incorrectly interpreted header len AND operation.

## 🔨 BC:

- 💔 Frame now for the internal (but public) operations receive additional `[]byte` slice to skip bound checks.

---

v3.1.4 (14.06.2021)
-------------------

## 🩹 Fixes:

- 🐛 Fix: Duplicated RPC error message: [PR](https://github.com/spiral/goridge/pull/129)

---

v3.1.3 (11.06.2021)
-------------------

## 🩹 Fixes:

- 🐛 Fix: Performance optimization fixes [PR](https://github.com/spiral/goridge/pull/127)

---

v3.1.2 (10.06.2021)
-------------------

## 🩹 Fixes:

- 🐛 Fix: Bug with incorrect proto interface casting which leads to a panic.

---

v3.1.1 (09.06.2021)
-------------------

## 🚀 New:

- ✏️ Standardise error operations in the whole project.
- ✏️ Remove `frame` to `byte` type aliasing in hot paths.

---

v3.1.0 (09.06.2021)
-------------------

## 👀 New:

- ✏️ Add support for the `PROTOBUF` in the RPC.

---

### v3.0.1 (06.02.2021)

- Move relay interface to the pkg folder.
- Separate CI into Linux, MacOS, Windows and Linters.
- Correct README.md (Installation and sample)
- Add PR template.
- Add dependabot.

### v3.0.0 (18.01.2021)

- New protocol (pkg/frame/frame.md)
- Various payloads support: msgpack, raw bytes, JSON, gob.
- New repo architecture.
- Shared memory [alfa].

### v2.4.6 (21.10.2020)

- Remove RawMessage check for the payload (@wolfy-j)
- golang version in go.mod bumped to 1.15

## v2.4.2 (19.05.2020)

- Add phpstan analyzer
- Fix code warnings from phpstan
- Improve Relay factory and SocketRelay
- Improve test coverage
- Performance improvements
- See the full milestone here: [link](https://github.com/spiral/goridge/milestone/5?closed=1)

## v2.4.0 (05.05.2020)

- More tests for PHP (@vvval)
- Upgrade PHP version to the 7.2 (currently minimum supported)
- Add new RelayInterface [link](https://github.com/spiral/goridge/pull/56/files#diff-85a3f483116946b4093f21ad855af4a8) (
  @vvval)
- See the full milestone here: [link](https://github.com/spiral/goridge/issues?q=is%3Aclosed+milestone%3A2.4.0)

## v2.3.1 (21.04.2020)

- Syscall usage optimized. Now the data is packing and sending via 1 (or 2 in some cases) send_socket calls, instead of
  2-3 (by @vvval)
- Unix sockets supported on windows (AF_UNIX) starting from OS Build 17056.
- Add the ability to define own relay with a codec (by @Reasno)

## v2.3.0 (23.03.2020)

- Replace std encoding/json package with the https://github.com/json-iterator/go
- Add BORS and GHA support
- golang modules updated to v2

## v2.2.1 (30.11.2019)

- Fixed too strict StreamRelay check by @tarampampam

## v2.2.0 (29.11.2019)

- Update travis to support go 1.13
- Update tests (errors handling, simplify)
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
