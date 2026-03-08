// Package rpc provides [net/rpc] codec implementations that use the goridge
// frame protocol as a transport.
//
// [Codec] is a server-side codec (implements [rpc.ServerCodec]) and
// [ClientCodec] is the corresponding client-side codec (implements
// [rpc.ClientCodec]). Both support multiple serialization formats selected
// per-call via frame flags: Proto (protobuf), JSON, Gob, and Raw byte slices.
//
// Frames and byte buffers are managed through [sync.Pool] to reduce
// allocations on the hot path.
package rpc
