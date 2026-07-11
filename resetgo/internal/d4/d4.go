// Package d4 implements the IEEE 1284.4 (Dot4 / D4) protocol used to talk to
// Epson printers over a bidirectional transport, ported from the Python
// ez-reset tool.
package d4

import (
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"ezreset/internal/control"
	"ezreset/internal/transport"
)

// Packet is a single D4 packet.
type Packet struct {
	PSID    int
	SSID    int
	Credit  int
	Control int
	Payload []byte
}

// Command is a D4 control command.
type Command int

const (
	CmdInit         Command = 0
	CmdOpenChannel  Command = 1
	CmdCloseChannel Command = 2
	CmdCredit       Command = 3
	CmdCreditReq    Command = 4
	CmdExit         Command = 8
	CmdGetSocketID  Command = 9
)

var errors = map[byte]string{
	0x80: "Malformed packet",
	0x81: "No credit",
	0x82: "Reply without command",
	0x83: "Packet too big",
	0x84: "Channel not open",
	0x85: "Unknown Result",
	0x86: "Credit overflow",
	0x87: "Bad command/reply",
}

// D4Error is a protocol-level error.
type D4Error struct {
	Msg string
}

func (e *D4Error) Error() string { return e.Msg }

// Channel is an open D4 socket.
type Channel struct {
	d4       *D4
	SSID     int
	PSID     int
	MTU      int
	txCredit int
	rxCredit int
	rxMax    int
	rxQueue  []Packet
}

// Open opens the channel and grants initial receive credits.
func (c *Channel) Open() error {
	c.d4.OpenChannel(c)
	c.d4.Credit(c, c.rxMax)
	c.rxCredit += c.rxMax
	return nil
}

// Close closes the channel.
func (c *Channel) Close() error {
	c.d4.CloseChannel(c)
	return nil
}

func (c *Channel) ensureCredit() {
	for c.txCredit < 1 {
		if granted, err := c.d4.CreditRequest(c, 0); err == nil && granted >= 1 {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// Write sends data over the channel, splitting into MTU-sized frames.
func (c *Channel) Write(data []byte, progress func(int)) error {
	for len(data) > 0 {
		control := 0
		payload := data
		if len(data) > c.MTU-6 {
			payload = data[:c.MTU-6]
		}
		control |= 2
		data = data[len(payload):]

		credit := c.rxMax - c.rxCredit
		if credit > 0xFF {
			credit = 0xFF
		}
		packet := Packet{PSID: c.PSID, SSID: c.SSID, Credit: credit, Control: control, Payload: payload}
		c.rxCredit += credit

		c.ensureCredit()
		if err := c.d4.WritePacket(c, packet); err != nil {
			return err
		}

		if progress != nil {
			progress(len(payload))
		}
	}
	return nil
}

// Read returns the next received packet.
func (c *Channel) Read() (Packet, error) {
	credit := c.rxMax - c.rxCredit
	if credit > 0xFF {
		c.d4.Credit(c, credit)
		c.rxCredit += credit
	}
	return c.d4.ReadPacket(c)
}

// D4 is the top-level protocol handler.
type D4 struct {
	transport transport.Transport
	channels  map[int]*Channel
}

// New creates a D4 handler and enters 1284.4 mode on the transport.
func New(t transport.Transport) (*D4, error) {
	d := &D4{
		transport: t,
		channels:  map[int]*Channel{0x00: {d4: nil, SSID: 0x00, txCredit: 1}},
	}
	d.channels[0x00].d4 = d

	// Drain any queued periodic status messages.
	if err := t.Drain(); err != nil {
		return nil, err
	}

	// Escape other modes, enter 1284.4 mode.
	if err := t.Write([]byte("\x00\x00\x00\x1b\x01@EJL 1284.4\n@EJL\n@EJL\n")); err != nil {
		return nil, err
	}
	if _, err := t.Read(8); err != nil {
		return nil, err
	}

	d.channels[0x00].d4 = d
	if err := d.Init(); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *D4) getFreePSID() (int, error) {
	for i := 0; i < 0x100; i++ {
		if _, ok := d.channels[i]; !ok {
			return i, nil
		}
	}
	return 0, &D4Error{Msg: "No free PSIDs to allocate to channel open."}
}

// WritePacket serializes and writes a packet to the transport.
func (d *D4) WritePacket(channel *Channel, packet Packet) error {
	length := 6 + len(packet.Payload)
	header := make([]byte, 6)
	binary.BigEndian.PutUint16(header[0:2], uint16(packet.PSID))
	binary.BigEndian.PutUint16(header[2:4], uint16(packet.SSID))
	binary.BigEndian.PutUint16(header[4:6], uint16(length))
	header[4] = byte(packet.Credit)
	header[5] = byte(packet.Control)

	data := append(header, packet.Payload...)
	log.Printf("> %x", data[:min(len(data), 0x100)])
	if err := d.transport.Write(data); err != nil {
		return err
	}
	channel.txCredit--
	return nil
}

// ReadPacket returns the next packet destined for the given channel.
func (d *D4) ReadPacket(channel *Channel) (Packet, error) {
	for len(channel.rxQueue) == 0 {
		if err := d.readNextPacket(); err != nil {
			return Packet{}, err
		}
	}
	p := channel.rxQueue[0]
	channel.rxQueue = channel.rxQueue[1:]
	return p, nil
}

func (d *D4) readNextPacket() error {
	header, err := d.transport.Read(6)
	if err != nil {
		return err
	}
	psid := int(header[0])
	ssid := int(header[1])
	length := int(binary.BigEndian.Uint16(header[2:4]))
	credit := int(header[4])
	control := int(header[5])
	log.Printf("< %x", header)

	payload, err := d.transport.Read(length - 6)
	if err != nil {
		return err
	}
	log.Printf("< %x", append(header, payload...))

	channel, ok := d.channels[psid]
	if !ok {
		log.Printf("Received packet for closed socket ID %d", psid)
		return nil
	}

	channel.txCredit += credit
	channel.rxCredit--
	channel.rxQueue = append(channel.rxQueue, Packet{
		PSID:    psid,
		SSID:    ssid,
		Credit:  credit,
		Control: control,
		Payload: payload,
	})
	return nil
}

// Command sends a D4 control command and returns the response payload.
func (d *D4) Command(command Command, payload []byte) ([]byte, error) {
	if command != CmdInit && command != CmdExit {
		if d.channels[0].txCredit < 1 {
			return nil, &D4Error{Msg: "no credit on control channel"}
		}
	}

	packet := Packet{PSID: 0x00, SSID: 0x00, Credit: 1, Control: 0x00, Payload: append([]byte{byte(command)}, payload...)}
	if err := d.WritePacket(d.channels[0], packet); err != nil {
		return nil, err
	}
	res, err := d.ReadPacket(d.channels[0])
	if err != nil {
		return nil, err
	}

	if res.PSID != 0 {
		return nil, &D4Error{Msg: fmt.Sprintf("unexpected PSID %d", res.PSID)}
	}

	if len(res.Payload) > 0 && res.Payload[0] == 0x7F {
		msg := errors[res.Payload[3]]
		if msg == "" {
			msg = fmt.Sprintf("0x%x", res.Payload[3])
		}
		log.Printf("D4 error: %s", msg)
	}

	if len(res.Payload) == 0 || res.Payload[0] != byte(command)|0x80 || res.Payload[1] != 0 {
		return nil, &D4Error{Msg: fmt.Sprintf("unexpected response %x", res.Payload)}
	}

	return res.Payload[2:], nil
}

// Init performs the D4 initialization handshake.
func (d *D4) Init() error {
	resp, err := d.Command(CmdInit, []byte{0x10})
	if err != nil {
		return err
	}
	if len(resp) != 1 || resp[0] != 0x10 {
		return &D4Error{Msg: fmt.Sprintf("Init: unexpected response %x", resp)}
	}
	return nil
}

// Exit leaves D4 mode.
func (d *D4) Exit() error {
	_, err := d.Command(CmdExit, nil)
	return err
}

// GetSocketID resolves a named socket to its SSID.
func (d *D4) GetSocketID(name string) (int, error) {
	resp, err := d.Command(CmdGetSocketID, []byte(name))
	if err != nil {
		return 0, err
	}
	return int(resp[0]), nil
}

// OpenChannel opens a channel for the given SSID.
func (d *D4) OpenChannel(channel *Channel) error {
	psid := channel.SSID
	req := make([]byte, 12)
	req[0] = byte(psid)
	req[1] = byte(channel.SSID)
	binary.BigEndian.PutUint16(req[2:4], 0xFFFF)
	binary.BigEndian.PutUint16(req[4:6], 0xFFFF)
	binary.BigEndian.PutUint16(req[6:8], 0x0000)
	binary.BigEndian.PutUint16(req[8:10], 0x0000)
	binary.BigEndian.PutUint16(req[10:12], 0x0000)

	res, err := d.Command(CmdOpenChannel, req)
	if err != nil {
		return err
	}
	rPSID, rSSID, mtu, _, credit := parseOpenChannelResp(res)
	if rSSID != channel.SSID {
		return &D4Error{Msg: fmt.Sprintf("OpenChannel: SSID mismatch %d != %d", rSSID, channel.SSID)}
	}
	channel.PSID = rPSID
	channel.MTU = mtu
	channel.txCredit = credit
	d.channels[rPSID] = channel
	return nil
}

func parseOpenChannelResp(res []byte) (psid, ssid, mtu, maxCredit, credit int) {
	psid = int(res[0])
	ssid = int(res[1])
	mtu = int(binary.BigEndian.Uint16(res[2:4]))
	maxCredit = int(binary.BigEndian.Uint16(res[4:6]))
	credit = int(binary.BigEndian.Uint16(res[6:8]))
	return
}

// CloseChannel closes an open channel.
func (d *D4) CloseChannel(channel *Channel) error {
	req := []byte{byte(channel.PSID), byte(channel.SSID)}
	if _, err := d.Command(CmdCloseChannel, req); err != nil {
		return err
	}
	delete(d.channels, channel.PSID)
	return nil
}

// Credit grants receive credits to a channel.
func (d *D4) Credit(channel *Channel, amount int) error {
	req := make([]byte, 4)
	req[0] = byte(channel.PSID)
	req[1] = byte(channel.SSID)
	binary.BigEndian.PutUint16(req[2:4], uint16(amount))
	_, err := d.Command(CmdCredit, req)
	return err
}

// CreditRequest requests transmit credits from the printer.
func (d *D4) CreditRequest(channel *Channel, amount int) (int, error) {
	if amount == 0 {
		amount = 0xFFFF
	}
	req := make([]byte, 4)
	req[0] = byte(channel.PSID)
	req[1] = byte(channel.SSID)
	binary.BigEndian.PutUint16(req[2:4], uint16(amount))
	resp, err := d.Command(CmdCreditReq, req)
	if err != nil {
		return 0, err
	}
	_, _, granted := parseCreditReqResp(resp)
	d.channels[channel.PSID].txCredit += granted
	return granted, nil
}

func parseCreditReqResp(resp []byte) (psid, ssid, amount int) {
	psid = int(resp[0])
	ssid = int(resp[1])
	amount = int(binary.BigEndian.Uint16(resp[2:4]))
	return
}

// Channel opens (or returns) a named EPSON control channel.
func (d *D4) Channel(name string) (*Channel, error) {
	ssid, err := d.GetSocketID(name)
	if err != nil {
		return nil, err
	}
	return &Channel{d4: d, SSID: ssid, rxMax: 0x0001}, nil
}

// ControlBackend adapts D4 to the control.ControlBackend interface.
type ControlBackend struct {
	transport transport.Transport
	d4        *D4
	channel   *Channel
}

// NewControlBackend creates a D4-based control backend.
func NewControlBackend(t transport.Transport) *ControlBackend {
	return &ControlBackend{transport: t}
}

// Open enters D4 mode and opens the EPSON-CTRL channel.
func (b *ControlBackend) Open() error {
	d4, err := New(b.transport)
	if err != nil {
		return err
	}
	b.d4 = d4
	ch, err := d4.Channel("EPSON-CTRL")
	if err != nil {
		return err
	}
	if err := ch.Open(); err != nil {
		return err
	}
	b.channel = ch
	return nil
}

// Close closes the channel.
func (b *ControlBackend) Close() error {
	if b.channel != nil {
		_ = b.channel.Close()
	}
	return nil
}

// Send writes a command and reads the response payload.
func (b *ControlBackend) Send(command []byte) ([]byte, error) {
	if b.channel == nil {
		return nil, &control.BackendError{Msg: "Channel must be opened"}
	}
	if err := b.channel.Write(command, nil); err != nil {
		return nil, err
	}
	res, err := b.channel.Read()
	if err != nil {
		return nil, err
	}
	return res.Payload, nil
}

// Identify returns the IEEE 1284.4 device ID.
func (b *ControlBackend) Identify() (string, error) {
	return b.transport.Identify()
}
