// Package pipe provides a pipe-based [relay.Relay] implementation that
// communicates with a child process over standard streams (STDIN/STDOUT).
//
// On Windows, prefer TCP sockets via the socket package for better reliability.
// A pipe relay closes automatically when the underlying process exits.
package pipe
