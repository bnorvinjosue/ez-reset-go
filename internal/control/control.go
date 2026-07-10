// Package control defines the ControlBackend interface used to send commands
// over the printer's control interface, ported from the Python ez-reset tool.
package control
//
// Control commands are always two bytes, followed by a little-endian length,
// then length bytes of payload. The payload is command-dependant and does not
// follow any common structure.
//
// Common commands are:
//   - "st": status information
//   - "vi": version information
//   - "||": service command
type ControlBackend interface {
	// Open prepares the backend for communication.
	Open() error
	// Close releases any resources held by the backend.
	Close() error
	// Send transmits the binary payload and returns the response.
	Send(command []byte) ([]byte, error)
	// Identify returns the IEEE 1284.4 device ID.
	Identify() (string, error)
}
