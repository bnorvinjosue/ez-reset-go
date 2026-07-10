//go:build !windows

// Package transport: stub USBPRINT implementation for non-Windows platforms.
// The real transport requires the Win32 USBPRINT API and only works on Windows.
// This stub lets the rest of the code compile and be tested on other platforms.
package transport

import "fmt"

// USBPRINTTransport is a non-functional placeholder on non-Windows platforms.
type USBPRINTTransport struct {
	Path   string
	closed bool
}

// NewUSBPRINTTransport creates a placeholder transport.
func NewUSBPRINTTransport(path string) *USBPRINTTransport {
	return &USBPRINTTransport{Path: path, closed: true}
}

// Open always returns an error on non-Windows platforms.
func (t *USBPRINTTransport) Open() error {
	return fmt.Errorf("USBPRINT transport is only supported on Windows")
}

// Close is a no-op on non-Windows platforms.
func (t *USBPRINTTransport) Close() error {
	t.closed = true
	return nil
}

// Closed reports whether the transport is closed.
func (t *USBPRINTTransport) Closed() bool {
	return t.closed
}

// Write always returns an error on non-Windows platforms.
func (t *USBPRINTTransport) Write(_ []byte) error {
	return fmt.Errorf("USBPRINT transport is only supported on Windows")
}

// Read always returns an error on non-Windows platforms.
func (t *USBPRINTTransport) Read(_ int) ([]byte, error) {
	return nil, fmt.Errorf("USBPRINT transport is only supported on Windows")
}

// Drain always returns an error on non-Windows platforms.
func (t *USBPRINTTransport) Drain() error {
	return fmt.Errorf("USBPRINT transport is only supported on Windows")
}

// Identify always returns an error on non-Windows platforms.
func (t *USBPRINTTransport) Identify() (string, error) {
	return "", fmt.Errorf("USBPRINT transport is only supported on Windows")
}
