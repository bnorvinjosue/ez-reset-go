#include "printer.h"
#include "utils.h"

#include <stdexcept>
#include <string>

namespace ezreset {

std::vector<unsigned char> Printer::sendCommand(const std::vector<unsigned char>& command,
                                                const std::vector<unsigned char>& payload) {
    std::vector<unsigned char> full = command;
    full.push_back(static_cast<unsigned char>(payload.size() & 0xFF));
    full.push_back(static_cast<unsigned char>((payload.size() >> 8) & 0xFF));
    full.insert(full.end(), payload.begin(), payload.end());
    return backend_->send(full);
}

std::vector<unsigned char> Printer::sendFactoryCommand(const std::vector<unsigned char>& model,
                                                       int action,
                                                       const std::vector<unsigned char>& payload) {
    std::vector<unsigned char> command = {'|', '|'};
    std::vector<unsigned char> actionCode = {
        static_cast<unsigned char>(action),
        static_cast<unsigned char>(action ^ 0xFF),
        static_cast<unsigned char>(((action >> 1) & 0x7F) | ((action << 7) & 0x80))};
    std::vector<unsigned char> full = model;
    full.insert(full.end(), actionCode.begin(), actionCode.end());
    full.insert(full.end(), payload.begin(), payload.end());
    return sendCommand(command, full);
}

Status Printer::getStatus() {
    std::vector<unsigned char> expected = {'@','B','D','C',' ','S','T','2','\r','\n'};
    auto response = sendCommand({'s','t'}, {'\x01'});
    if (response.size() < expected.size()
        || std::string(response.begin(), response.begin() + expected.size())
               != std::string(expected.begin(), expected.end())) {
        throw std::runtime_error("unknown status response");
    }
    std::vector<unsigned char> payload(response.begin() + expected.size(), response.end());
    return statusFromBytes(payload);
}

int Printer::readEEPROM(int address) {
    int action = 0x41;
    std::vector<unsigned char> expected = {'@','B','D','C',' ','P','S','\r','\n'};
    std::vector<unsigned char> addr = {
        static_cast<unsigned char>(address & 0xFF),
        static_cast<unsigned char>((address >> 8) & 0xFF)};
    auto response = sendFactoryCommand(device_.model, action, addr);
    if (response.size() < expected.size()
        || std::string(response.begin(), response.begin() + expected.size())
               != std::string(expected.begin(), expected.end())) {
        throw std::runtime_error("unknown EEPROM response");
    }
    std::string hex(response.begin() + 16, response.begin() + 18);
    return static_cast<int>(std::stoul(hex, nullptr, 16));
}

std::vector<unsigned char> Printer::readEEPROMMultiple(const std::vector<int>& addresses) {
    std::vector<unsigned char> out;
    for (int a : addresses) out.push_back(static_cast<unsigned char>(readEEPROM(a)));
    return out;
}

std::vector<unsigned char> Printer::readEEPROMRange(int address, int size) {
    int action = 0x51;
    std::vector<unsigned char> expected = {'@','B','D','C',' ','P','S','\r','\n'};
    std::vector<unsigned char> payload = {
        static_cast<unsigned char>(address & 0xFF),
        static_cast<unsigned char>((address >> 8) & 0xFF),
        static_cast<unsigned char>(size & 0xFF)};
    auto response = sendFactoryCommand(device_.model, action, payload);
    if (response.size() < expected.size()
        || std::string(response.begin(), response.begin() + expected.size())
               != std::string(expected.begin(), expected.end())) {
        throw std::runtime_error("unknown EEPROM range response");
    }
    std::string hex(response.begin() + 16, response.begin() + 16 + size * 2);
    std::vector<unsigned char> out;
    for (size_t i = 0; i + 1 < hex.size(); i += 2) {
        out.push_back(static_cast<unsigned char>(std::stoul(hex.substr(i, 2), nullptr, 16)));
    }
    return out;
}

void Printer::writeEEPROM(int address, int value) {
    int action = 0x42;
    std::vector<unsigned char> payload = {
        static_cast<unsigned char>(address & 0xFF),
        static_cast<unsigned char>((address >> 8) & 0xFF),
        static_cast<unsigned char>(value & 0xFF)};
    payload.insert(payload.end(), device_.key.begin(), device_.key.end());
    sendFactoryCommand(device_.model, action, payload);
}

std::vector<std::pair<int, int>> Printer::getWaste() {
    std::vector<std::pair<int, int>> result;
    for (const auto& counter : device_.counters) {
        auto value = readEEPROMMultiple(counter.addresses);
        int v = 0;
        for (size_t i = 0; i < value.size(); i++) {
            v |= (static_cast<int>(value[i]) << (8 * i));
        }
        result.push_back({v, counter.max});
    }
    return result;
}

void Printer::resetWaste() {
    for (const auto& kv : device_.reset) {
        writeEEPROM(kv.first, kv.second);
    }
}

std::string Printer::getSerial() {
    return getStatus().serial;
}

void Printer::clean(int level) {
    sendFactoryCommand(device_.model, 0x84, {static_cast<unsigned char>(level & 0xFF)});
}

void Printer::powerOff() {
    sendFactoryCommand(device_.model, 0x20, {});
}

void Printer::restart() {
    sendFactoryCommand(device_.model, 0x21, {});
}

} // namespace ezreset
