#ifndef EZRESET_PRINTER_H
#define EZRESET_PRINTER_H

#include <string>
#include <vector>
#include <utility>

#include "devices.h"
#include "status.h"

namespace ezreset {

// High-level printer operations on top of a control backend.
class Printer {
public:
    Printer(class ControlBackend* backend, const Device& device)
        : backend_(backend), device_(device) {}

    Status getStatus();
    std::vector<std::pair<int, int>> getWaste();
    void resetWaste();
    std::string getSerial();
    void clean(int level);
    void powerOff();
    void restart();

private:
    std::vector<unsigned char> sendCommand(const std::vector<unsigned char>& command,
                                           const std::vector<unsigned char>& payload);
    std::vector<unsigned char> sendFactoryCommand(const std::vector<unsigned char>& model,
                                                  int action,
                                                  const std::vector<unsigned char>& payload);
    int readEEPROM(int address);
    std::vector<unsigned char> readEEPROMMultiple(const std::vector<int>& addresses);
    std::vector<unsigned char> readEEPROMRange(int address, int size);
    void writeEEPROM(int address, int value);

    class ControlBackend* backend_;
    Device device_;
};

// Minimal control backend interface (implemented by D4ControlBackend).
class ControlBackend {
public:
    virtual ~ControlBackend() = default;
    virtual void open() = 0;
    virtual void close() = 0;
    virtual std::vector<unsigned char> send(const std::vector<unsigned char>& command) = 0;
    virtual std::string identify() = 0;
};

} // namespace ezreset

#endif // EZRESET_PRINTER_H
