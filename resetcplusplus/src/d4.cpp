#include "d4.h"

#include <chrono>
#include <thread>
#include <cstring>
#include <iostream>

namespace ezreset {

namespace {
const char* d4ErrorText(unsigned char code) {
    switch (code) {
        case 0x80: return "Malformed packet";
        case 0x81: return "No credit";
        case 0x82: return "Reply without command";
        case 0x83: return "Packet too big";
        case 0x84: return "Channel not open";
        case 0x85: return "Unknown Result";
        case 0x86: return "Credit overflow";
        case 0x87: return "Bad command/reply";
        default: return "Unknown D4 error";
    }
}
} // namespace

D4::D4(Transport* transport) : transport_(transport) {
    channels_[0x00] = new D4Channel(this, 0x00);
    channels_[0x00]->setTxCredit(1);

    transport_->drain();

    std::vector<unsigned char> enter = {
        '\x00','\x00','\x00','\x1b','\x01','@','E','J','L',' ','1','2','8','4','.','4',
        '\n','@','E','J','L','\n','@','E','J','L','\n'};
    transport_->write(enter);
    transport_->read(8);

    init();
}

D4::~D4() {
    for (auto& kv : channels_) delete kv.second;
}

void D4Channel::open() {
    d4_->openChannel(this);
    d4_->credit(this, rxMax_);
    rxCredit_ += rxMax_;
}

void D4Channel::close() {
    d4_->closeChannel(this);
}

void D4Channel::write(const std::vector<unsigned char>& data) {
    std::vector<unsigned char> remaining = data;
    while (!remaining.empty()) {
        int control = 0;
        std::vector<unsigned char> payload = remaining;
        if (static_cast<int>(remaining.size()) > mtu_ - 6) {
            payload.assign(remaining.begin(), remaining.begin() + (mtu_ - 6));
        }
        control |= 2;
        remaining.erase(remaining.begin(), remaining.begin() + payload.size());

        int credit = rxMax_ - rxCredit_;
        if (credit > 0xFF) credit = 0xFF;
        D4Packet packet{psid_, ssid_, credit, control, payload};
        rxCredit_ += credit;

        ensureCredit();
        d4_->writePacket(this, packet);
    }
}

void D4Channel::ensureCredit() {
    while (txCredit_ < 1) {
        if (d4_->creditRequest(this) >= 1) return;
        std::this_thread::sleep_for(std::chrono::milliseconds(100));
    }
}

D4Packet D4Channel::read() {
    int credit = rxMax_ - rxCredit_;
    if (credit > 0xFF) {
        d4_->credit(this, credit);
        rxCredit_ += credit;
    }
    return d4_->readPacket(this);
}

void D4::writePacket(D4Channel* channel, const D4Packet& packet) {
    int length = 6 + static_cast<int>(packet.payload.size());
    std::vector<unsigned char> header(6);
    header[0] = static_cast<unsigned char>(packet.psid);
    header[1] = static_cast<unsigned char>(packet.ssid);
    header[2] = static_cast<unsigned char>((length >> 8) & 0xFF);
    header[3] = static_cast<unsigned char>(length & 0xFF);
    header[4] = static_cast<unsigned char>(packet.credit);
    header[5] = static_cast<unsigned char>(packet.control);

    std::vector<unsigned char> data = header;
    data.insert(data.end(), packet.payload.begin(), packet.payload.end());
    transport_->write(data);

    channel->addTxCredit(-1);
}

void D4::readNextPacket() {
    std::vector<unsigned char> header = transport_->read(6);
    int psid = header[0];
    int ssid = header[1];
    int length = (header[2] << 8) | header[3];
    int credit = header[4];
    int control = header[5];
    std::vector<unsigned char> payload = transport_->read(length - 6);

    auto it = channels_.find(psid);
    if (it == channels_.end()) {
        std::cerr << "Received packet for closed socket ID " << psid << "\n";
        return;
    }
    D4Channel* channel = it->second;
    channel->addTxCredit(credit);
    channel->addRxCredit(-1);
    channel->pushQueue({psid, ssid, credit, control, payload});
}

D4Packet D4::readPacket(D4Channel* channel) {
    while (channel->queueEmpty()) {
        readNextPacket();
    }
    return channel->popQueue();
}

std::vector<unsigned char> D4::command(D4Command cmd,
                                       const std::vector<unsigned char>& payload) {
    if (cmd != D4Command::Init && cmd != D4Command::Exit) {
        if (channels_[0]->txCredit() < 1) {
            throw D4Error("no credit on control channel");
        }
    }
    std::vector<unsigned char> full;
    full.push_back(static_cast<unsigned char>(cmd));
    full.insert(full.end(), payload.begin(), payload.end());

    writePacket(channels_[0x00], {0x00, 0x00, 1, 0x00, full});
    D4Packet res = readPacket(channels_[0x00]);
    if (res.psid != 0) throw D4Error("unexpected PSID");

    if (!res.payload.empty() && res.payload[0] == 0x7F) {
        std::cerr << "D4 error: " << d4ErrorText(res.payload[3]) << "\n";
    }
    if (res.payload.empty() || res.payload[0] != (static_cast<int>(cmd) | 0x80)
        || res.payload[1] != 0) {
        throw D4Error("unexpected D4 response");
    }
    return std::vector<unsigned char>(res.payload.begin() + 2, res.payload.end());
}

void D4::init() {
    auto resp = command(D4Command::Init, {'\x10'});
    if (resp.size() != 1 || resp[0] != 0x10) throw D4Error("Init: unexpected response");
}

void D4::exit() {
    command(D4Command::Exit);
}

int D4::getSocketID(const std::string& name) {
    auto resp = command(D4Command::GetSocketID,
                        std::vector<unsigned char>(name.begin(), name.end()));
    return resp[0];
}

int D4::getFreePSID() {
    for (int i = 0; i < 0x100; i++) {
        if (channels_.find(i) == channels_.end()) return i;
    }
    throw D4Error("No free PSIDs to allocate to channel open.");
}

void D4::openChannel(D4Channel* channel) {
    int psid = channel->ssid();
    std::vector<unsigned char> req(12);
    req[0] = static_cast<unsigned char>(psid);
    req[1] = static_cast<unsigned char>(channel->ssid());
    req[2] = 0xFF; req[3] = 0xFF;
    req[4] = 0xFF; req[5] = 0xFF;
    req[10] = 0x00; req[11] = 0x00;

    auto res = command(D4Command::OpenChannel, req);
    int rPsid = res[0];
    int rSsid = res[1];
    int mtu = (res[2] << 8) | res[3];
    int credit = (res[6] << 8) | res[7];
    if (rSsid != channel->ssid()) throw D4Error("OpenChannel: SSID mismatch");
    channel->setPsid(rPsid);
    channel->setMtu(mtu);
    channel->setTxCredit(credit);
    channels_[rPsid] = channel;
}

void D4::closeChannel(D4Channel* channel) {
    std::vector<unsigned char> req = {
        static_cast<unsigned char>(channel->psid()),
        static_cast<unsigned char>(channel->ssid())};
    command(D4Command::CloseChannel, req);
    channels_.erase(channel->psid());
}

void D4::credit(D4Channel* channel, int amount) {
    std::vector<unsigned char> req = {
        static_cast<unsigned char>(channel->psid()),
        static_cast<unsigned char>(channel->ssid()),
        static_cast<unsigned char>((amount >> 8) & 0xFF),
        static_cast<unsigned char>(amount & 0xFF)};
    command(D4Command::Credit, req);
}

int D4::creditRequest(D4Channel* channel, int amount) {
    if (amount == 0) amount = 0xFFFF;
    std::vector<unsigned char> req = {
        static_cast<unsigned char>(channel->psid()),
        static_cast<unsigned char>(channel->ssid()),
        static_cast<unsigned char>((amount >> 8) & 0xFF),
        static_cast<unsigned char>(amount & 0xFF)};
    auto resp = command(D4Command::CreditRequest, req);
    int granted = (resp[2] << 8) | resp[3];
    channels_[channel->psid()]->addTxCredit(granted);
    return granted;
}

D4Channel* D4::channel(const std::string& name) {
    int ssid = getSocketID(name);
    return new D4Channel(this, ssid);
}

// ---- D4ControlBackend ----

D4ControlBackend::~D4ControlBackend() {
    close();
}

void D4ControlBackend::open() {
    d4_ = new D4(transport_);
    channel_ = d4_->channel("EPSON-CTRL");
    channel_->open();
}

void D4ControlBackend::close() {
    if (channel_) { channel_->close(); delete channel_; channel_ = nullptr; }
    if (d4_) { delete d4_; d4_ = nullptr; }
}

std::vector<unsigned char> D4ControlBackend::send(const std::vector<unsigned char>& command) {
    if (!channel_) throw std::runtime_error("Channel must be opened");
    channel_->write(command);
    D4Packet res = channel_->read();
    return res.payload;
}

std::string D4ControlBackend::identify() {
    return transport_->identify();
}

} // namespace ezreset
