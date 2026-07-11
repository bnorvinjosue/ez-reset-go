#ifndef EZRESET_UTILS_H
#define EZRESET_UTILS_H

#include <string>
#include <vector>
#include <map>

namespace ezreset {

// A single entry inside a binary status struct.
struct StatusEntry {
    unsigned char header;
    unsigned char length;
    std::vector<unsigned char> payload;
};

// Parse a binary status struct into its entries.
// The struct is: [u16 little-endian length][entry...] where each entry is
// [header][size][payload...].
std::vector<StatusEntry> parseStatusStruct(const std::vector<unsigned char>& data);

// Parse an IEEE 1284.4 ID string ("KEY:VAL;KEY:VAL;...") into a map.
std::map<std::string, std::string> parseIdentifier(const std::string& identifier);

// Extract a single field (e.g. "MDL") from an IEEE 1284 ID string.
std::string parseField(const std::string& identifier, const std::string& key);

} // namespace ezreset

#endif // EZRESET_UTILS_H
