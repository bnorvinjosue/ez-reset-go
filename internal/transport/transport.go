// Package transport defines the bidirectional transport interface used by the
// D4 and END4 control backends, ported from the Python ez-reset tool.
package transport

// Transport is a bidirectional channel to a printer. Implementations must be
// safe to use as a context manager: call Open before use and Close afterwards.
type Transport interface {
	// Open prepares the transport for communication.
	Open() error
	// Close releases any resources held by the transport.
	Close() error
	// Closed reports whether the transport is currently closed.
	Closed() bool
	// Write sends raw bytes to the printer.
	Write(data []byte) error
	// Read blocks until exactly size bytes have been read.
	Read(size int) ([]byte, error)
	// Drain discards any pending input.
	Drain() error
	// Identify returns the IEEE 1284.4 device ID string.
	Identify() (string, error)
}
