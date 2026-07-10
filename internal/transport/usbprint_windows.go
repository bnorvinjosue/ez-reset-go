//go:build windows

// Package transport: Windows USBPRINT implementation using the Win32 API via
// golang.org/x/sys/windows, ported from the Python ez-reset tool.
package transport

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

const maxTransferSize = 0x400000

// USBPRINTTransport talks to an Epson printer over the USBPRINT interface on
// Windows using CreateFileW / DeviceIoControl.
type USBPRINTTransport struct {
	Path   string
	handle windows.Handle
	closed bool
	buffer []byte
}

// NewUSBPRINTTransport creates a transport for the given device path
// (e.g. \\.\USBPRINT\...).
func NewUSBPRINTTransport(path string) *USBPRINTTransport {
	return &USBPRINTTransport{Path: path, closed: true}
}

// Open opens a handle to the USBPRINT device and issues a soft reset.
func (t *USBPRINTTransport) Open() error {
	pathUTF16, err := windows.UTF16PtrFromString(t.Path)
	if err != nil {
		return err
	}

	handle, err := windows.CreateFile(
		pathUTF16,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_NO_BUFFERING|windows.FILE_FLAG_WRITE_THROUGH,
		0,
	)
	if err != nil {
		return fmt.Errorf("CreateFileW(%s): %w", t.Path, err)
	}

	t.handle = handle

	// Issue a soft reset.
	if _, err := t.deviceIoControl(ioctlUsbprintSoftReset, nil, 1024); err != nil {
		windows.CloseHandle(t.handle)
		return fmt.Errorf("soft reset: %w", err)
	}

	t.closed = false
	return nil
}

// Close closes the device handle.
func (t *USBPRINTTransport) Close() error {
	if t.closed {
		return nil
	}
	if err := windows.CloseHandle(t.handle); err != nil {
		return err
	}
	t.closed = true
	return nil
}

// Closed reports whether the transport is closed.
func (t *USBPRINTTransport) Closed() bool {
	return t.closed
}

// Write sends raw bytes to the printer.
func (t *USBPRINTTransport) Write(data []byte) error {
	if t.closed {
		return fmt.Errorf("handle to USBPRINT device %s is closed", t.Path)
	}

	written := uint32(0)
	if err := windows.WriteFile(t.handle, data, &written, nil); err != nil {
		return err
	}
	if int(written) != len(data) {
		return fmt.Errorf("wrote %d of %d bytes", written, len(data))
	}
	return nil
}

// Read blocks until exactly size bytes have been read.
func (t *USBPRINTTransport) Read(size int) ([]byte, error) {
	if t.closed {
		return nil, fmt.Errorf("handle to USBPRINT device %s is closed", t.Path)
	}

	for len(t.buffer) < size {
		buf := make([]byte, maxTransferSize)
		read := uint32(0)
		if err := windows.ReadFile(t.handle, buf, &read, nil); err != nil {
			return nil, err
		}
		t.buffer = append(t.buffer, buf[:read]...)
	}

	out := t.buffer[:size]
	t.buffer = t.buffer[size:]
	return out, nil
}

// Drain discards any pending input.
func (t *USBPRINTTransport) Drain() error {
	for {
		buf := make([]byte, maxTransferSize)
		read := uint32(0)
		if err := windows.ReadFile(t.handle, buf, &read, nil); err != nil {
			return err
		}
		if read == 0 {
			return nil
		}
	}
}

// Identify returns the IEEE 1284.4 device ID string.
func (t *USBPRINTTransport) Identify() (string, error) {
	out, err := t.deviceIoControl(ioctlUsbprintGet1284ID, nil, 1024)
	if err != nil {
		return "", err
	}
	// The first two bytes are a little-endian length prefix.
	if len(out) < 2 {
		return "", fmt.Errorf("identify: short response")
	}
	_ = binary.LittleEndian.Uint16(out[0:2])
	return string(out[2:]), nil
}

// ensure binary is used (kept for potential length-prefix handling).
var _ = binary.LittleEndian

func (t *USBPRINTTransport) deviceIoControl(ioctl uint32, in []byte, outSize int) ([]byte, error) {
	out := make([]byte, outSize)
	var bytesReturned uint32

	inPtr := (*byte)(unsafe.Pointer(nil))
	if len(in) > 0 {
		inPtr = &in[0]
	}

	err := windows.DeviceIoControl(
		t.handle,
		ioctl,
		inPtr,
		uint32(len(in)),
		&out[0],
		uint32(outSize),
		&bytesReturned,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return out[:bytesReturned], nil
}

// IOCTL codes for the USBPRINT driver.
const (
	ioctlUsbprintGet1284ID = 2228276
	ioctlUsbprintSoftReset = 2228288
)
