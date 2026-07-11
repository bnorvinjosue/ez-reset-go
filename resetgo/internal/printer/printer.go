// Package printer implements the high-level printer operations (status, EEPROM
// read/write, waste counter reset) on top of a control backend, ported from the
// Python ez-reset tool.
package printer

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"ezreset/internal/control"
	"ezreset/internal/devices"
	"ezreset/internal/status"
	"ezreset/internal/utils"
)

// Printer wraps a control backend and a device definition.
type Printer struct {
	Device  devices.Device
	control control.ControlBackend
}

// New creates a Printer.
func New(backend control.ControlBackend, device devices.Device) *Printer {
	return &Printer{Device: device, control: backend}
}

func (p *Printer) sendCommand(command, payload []byte) ([]byte, error) {
	full := append(command, byte(len(payload)&0xFF), byte((len(payload)>>8)&0xFF))
	full = append(full, payload...)
	return p.control.Send(full)
}

func (p *Printer) sendFactoryCommand(model []byte, action int, payload []byte) ([]byte, error) {
	command := []byte("||")
	actionCode := []byte{
		byte(action),
		byte(action ^ 0xFF),
		byte(((action >> 1) & 0x7F) | ((action << 7) & 0x80)),
	}
	full := append(model, actionCode...)
	full = append(full, payload...)
	return p.sendCommand(command, full)
}

// GetStatus returns the parsed printer status.
func (p *Printer) GetStatus() (*status.Status, error) {
	expected := []byte("@BDC ST2\r\n")
	response, err := p.sendCommand([]byte("st"), []byte{0x01})
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(string(response), string(expected)) {
		return nil, fmt.Errorf("unknown response %q for command %q", response, expected)
	}
	payload := response[len(expected):]
	return status.StatusFromBytes(payload)
}

func (p *Printer) readEEPROM(address int) (int, error) {
	action := 0x41
	expected := []byte("@BDC PS\r\n")
	response, err := p.sendFactoryCommand(p.Device.Model, action, uint16Bytes(address))
	if err != nil {
		return 0, err
	}
	if !strings.HasPrefix(string(response), string(expected)) {
		return 0, fmt.Errorf("unknown response %q for command %q", response, expected)
	}
	return parseIntHex(string(response[16:18])), nil
}

func (p *Printer) writeEEPROM(address, value int) ([]byte, error) {
	action := 0x42
	payload := append(uint16Bytes(address), byte(value))
	payload = append(payload, p.Device.Key...)
	return p.sendFactoryCommand(p.Device.Model, action, payload)
}

func (p *Printer) readEEPROMMultiple(addresses []int) ([]byte, error) {
	out := make([]byte, 0, len(addresses))
	for _, a := range addresses {
		v, err := p.readEEPROM(a)
		if err != nil {
			return nil, err
		}
		out = append(out, byte(v))
	}
	return out, nil
}

func (p *Printer) readEEPROMRange(address, size int) ([]byte, error) {
	action := 0x51
	expected := []byte("@BDC PS\r\n")
	payload := append(uint16Bytes(address), byte(size))
	response, err := p.sendFactoryCommand(p.Device.Model, action, payload)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(string(response), string(expected)) {
		return nil, fmt.Errorf("unknown response %q for command %q", response, expected)
	}
	return hex.DecodeString(strings.TrimSpace(string(response[16 : 16+size*2])))
}

// Identify returns the parsed IEEE 1284.4 identifier fields.
func (p *Printer) Identify() (map[string]string, error) {
	id, err := p.control.Identify()
	if err != nil {
		return nil, err
	}
	return utils.ParseIdentifier(id), nil
}

// GetSerial returns the printer serial number.
func (p *Printer) GetSerial() (string, error) {
	st, err := p.GetStatus()
	if err != nil {
		return "", err
	}
	return st.Serial, nil
}

// GetWaste returns the current waste counter values and their maxima.
func (p *Printer) GetWaste() ([][2]int, error) {
	var result [][2]int
	for _, counter := range p.Device.Counters {
		value, err := p.readEEPROMMultiple(counter.Addresses)
		if err != nil {
			return nil, err
		}
		result = append(result, [2]int{int(binary.LittleEndian.Uint16(value)), counter.Max})
	}
	return result, nil
}

// ResetWaste writes the reset values for all waste counters.
func (p *Printer) ResetWaste() error {
	for addr, value := range p.Device.Reset {
		if _, err := p.writeEEPROM(addr, value); err != nil {
			return err
		}
	}
	return nil
}

// Clean triggers a cleaning cycle at the given level.
func (p *Printer) Clean(level int) error {
	_, err := p.sendFactoryCommand(p.Device.Model, 0x84, []byte{byte(level)})
	return err
}

// PowerOff powers the printer off.
func (p *Printer) PowerOff() error {
	_, err := p.sendFactoryCommand(p.Device.Model, 0x20, nil)
	return err
}

// Restart restarts the printer.
func (p *Printer) Restart() error {
	_, err := p.sendFactoryCommand(p.Device.Model, 0x21, nil)
	return err
}

func uint16Bytes(v int) []byte {
	return []byte{byte(v & 0xFF), byte((v >> 8) & 0xFF)}
}

func parseIntHex(s string) int {
	v, _ := hex.DecodeString(strings.TrimSpace(s))
	if len(v) == 0 {
		return 0
	}
	return int(v[0])
}
