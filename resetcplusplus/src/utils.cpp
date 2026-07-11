#include "utils.h"

#include <stdexcept>
#include <sstream>

namespace ezreset {

std::vector<StatusEntry> parseStatusStruct(const std::vector<unsigned char>& data) {
    if (data.size() < 2) {
        throw std::runtime_error("status payload too short");
    }
    unsigned int length = data[0] | (static_cast<unsigned int>(data[1]) << 8);
    if (data.size() != length + 2) {
        throw std::runtime_error("status payload length invalid");
    }

    std::vector<StatusEntry> entries;
    size_t index = 2;
    while (index < length) {
        StatusEntry entry;
        entry.header = data[index++];
        entry.length = data[index++];
        if (index + entry.length > data.size()) {
            throw std::runtime_error("status entry truncated");
        }
        entry.payload.assign(data.begin() + index, data.begin() + index + entry.length);
        index += entry.length;
        entries.push_back(entry);
    }
    return entries;
}

std::map<std::string, std::string> parseIdentifier(const std::string& identifier) {
    std::map<std::string, std::string> result;
    std::stringstream ss(identifier);
    std::string field;
    while (std::getline(ss, field, ';')) {
        if (field.empty()) continue;
        auto pos = field.find(':');
        if (pos != std::string::npos) {
            result[field.substr(0, pos)] = field.substr(pos + 1);
        }
    }
    return result;
}

std::string parseField(const std::string& identifier, const std::string& key) {
    auto map = parseIdentifier(identifier);
    auto it = map.find(key);
    return it != map.end() ? it->second : std::string();
}

} // namespace ezreset
