// Package socket provides a socket-based [relay.Relay] implementation for
// frame transport over TCP and Unix domain connections.
//
// It wraps any [io.ReadWriteCloser] (typically a [net.Conn]) and uses
// pre-allocated buffer pools from the internal package to minimize allocations
// during frame reception.
package socket
