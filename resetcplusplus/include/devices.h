#ifndef EZRESET_DEVICES_H
#define EZRESET_DEVICES_H

#include <string>
#include <vector>
#include <map>

namespace ezreset {

struct Counter {
    std::vector<int> addresses;
    int max = 0;
};

struct Device {
    std::vector<unsigned char> model;
    std::vector<unsigned char> key;
    std::vector<Counter> counters;
    std::map<int, int> reset;
};

// Load the device database from devices.xml at the given path.
std::map<std::string, Device> loadDevices(const std::string& path);

// Resolve a device definition for the given model string.
Device deviceByModel(const std::map<std::string, Device>& devices, const std::string& model);

} // namespace ezreset

#endif // EZRESET_DEVICES_H
