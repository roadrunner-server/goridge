```log
    0                   1                   2                     3   OCTET
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |Version|   HL  |     Flags     |         Payload Length        |    0
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |        Payload Length         |           Header CRC          |    4
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |          Header CRC           |           Reserved            |    8
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |           Options             |            Payload            |    12
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

1. `0-th` byte contains version and header length (HL). HL calculated in 32bit words. For example, HL is 3, that means, that size of the header is 3*32bit = 96bits = 12 bytes.
2. `1-st` byte contains flags. The flags described in frame_flags.go file. It consists of overlapping and non-overlapping flags.
Overlapping flags are just bit flags. They might be combined with bitwise OR and checked with bitwise AND. Non-overlapping flags
   can't be used with other flags. In means, that if you have non-overlapping flag in 1-st byte, you can't use other flags.
   
3. `(2, 3, 4, 5)` bytes contain payload length and presented by unsigned long 32bit integer (up to 4Gb in payload).
4. `(6, 7, 8, 9)` bytes contain header `CRC32` checksum. CRC32 calculated only for `0-5` (including) bytes.
5. `(10, 11)` bytes are padding and currently reserved for future use. For example as `sequence Id` for async operations.
6. `(12..52)` bytes contain options. Options are optional. As an example of usage, in `goridge` in case of pipes or sockets
we write two unsigned 32bit integers of RPC_SEQ_ID and method length offset. This field can be up to 40 bytes.
   
7. `From (12..52)` lays payload. Maximum payload, that can be transmitted via 1 frame is `4Gb`.