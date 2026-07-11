// Package end4 implements the Epson-proprietary END4 protocol, which makes it
// possible to send CTRL commands over a printer data line without requiring
// full IEEE 1284.4 framing, ported from the Python ez-reset tool.
package end4

import (
	"fmt"
	"time"

	"ezreset/internal/control"
	"ezreset/internal/transport"
	"ezreset/internal/utils"
)

var exitPacketMode2 = []byte("\x00\x00\x00\x1b\x01@EJL 1284.4\n@EJL\t\t\t\t\t\n")

// ControlBackend handles the Control channel over END4.
type ControlBackend struct {
	transport transport.Transport
}

// NewControlBackend creates an END4-based control backend.
func NewControlBackend(t transport.Transport) *ControlBackend {
	return &ControlBackend{transport: t}
}

// Open enters END4 mode by draining the data stream.
func (b *ControlBackend) Open() error {
	if b.transport.Closed() {
		return control.NewBackendError("BiDi device is closed")
	}

	identifier := utils.ParseIdentifier(b.identify())
	if err := b.transport.Write(exitPacketMode2); err != nil {
		return err
	}

	dds := 0
	if v, ok := identifier["DDS"]; ok {
		var parsed int
		if _, err := fmt.Sscanf(v, "%x", &parsed); err == nil {
			dds = parsed
		}
	}
	for dds > 0 {
		chunk := dds
		if chunk > 0x8000 {
			chunk = 0x8000
		}
		if err := b.transport.Write(bytesRepeat(0x11, chunk)); err != nil {
			return err
		}
		dds -= chunk
	}
	return nil
}

// Close is a no-op for END4.
func (b *ControlBackend) Close() error {
	return nil
}

// Send transmits a command over END4 and returns the response payload.
func (b *ControlBackend) Send(command []byte) ([]byte, error) {
	if err := b.transport.Drain(); err != nil {
		return nil, err
	}

	frame := make([]byte, 0, len(command)+14)
	frame = append(frame, []byte("END4")...)
	frame = append(frame, 0x02, 0x01, 0x00, 0x00, 0x00)
	frame = append(frame, byte(len(command)+14))
	frame = append(frame, 0x00, 0x00, 0x02, 0x00)
	frame = append(frame, command...)

	if err := b.transport.Write(frame); err != nil {
		return nil, err
	}

	response := []byte{}
	for {
		chunk, err := b.transport.Read(1024)
		if err != nil {
			return nil, err
		}
		response = append(response, chunk...)
		if len(response) >= 10 && string(response[0:4]) == "END4" {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	expectedLen := int(response[9])
	if len(response) != expectedLen {
		return nil, control.NewBackendError("Received incomplete packet.")
	}

	return response[10:], nil
}

// Identify returns the IEEE 1284.4 device ID.
func (b *ControlBackend) Identify() (string, error) {
	return b.identify(), nil
}

func (b *ControlBackend) identify() string {
	id, _ := b.transport.Identify()
	return id
}

func bytesRepeat(b byte, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = b
	}
	return out
}
