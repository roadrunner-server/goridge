// Package relay defines the Relay interface for frame-based inter-process
// communication.
//
// A Relay transports [frame.Frame] values between processes via three methods:
// Send, Receive, and Close. Concrete implementations are provided by the
// [github.com/roadrunner-server/goridge/v4/pkg/socket] and
// [github.com/roadrunner-server/goridge/v4/pkg/pipe] packages.
package relay
