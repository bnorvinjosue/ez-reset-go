#ifndef EZRESET_TRANSPORT_H
#define EZRESET_TRANSPORT_H

#include <string>
#include <vector>
#include <stdexcept>

namespace ezreset {

// A bidirectional channel to a printer. Implementations must be opened before
// use and closed afterwards.
class Transport {
public:
    virtual ~Transport() = default;
    virtual void open() = 0;
    virtual void close() = 0;
    virtual bool isClosed() const = 0;
    virtual void write(const std::vector<unsigned char>& data) = 0;
    virtual std::vector<unsigned char> read(int size) = 0;
    virtual void drain() = 0;
    virtual std::string identify() = 0;
};

// USBPRINT transport. On Windows it uses CreateFileW / DeviceIoControl.
// On other platforms it is a non-functional placeholder (the real transport
// requires the Win32 USBPRINT API).
class USBPRINTTransport : public Transport {
public:
    explicit USBPRINTTransport(const std::string& path);
    ~USBPRINTTransport() override;

    void open() override;
    void close() override;
    bool isClosed() const override { return closed_; }
    void write(const std::vector<unsigned char>& data) override;
    std::vector<unsigned char> read(int size) override;
    void drain() override;
    std::string identify() override;

private:
    std::string path_;
    bool closed_ = true;
#ifdef _WIN32
    void* handle_ = nullptr; // HANDLE
    std::vector<unsigned char> buffer_;
#endif
};

// Enumerate connected USBPRINT printers (Windows only).
std::vector<std::string> enumeratePrinters();

} // namespace ezreset

#endif // EZRESET_TRANSPORT_H
