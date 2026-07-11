#include "transport.h"

#include <stdexcept>

namespace ezreset {

USBPRINTTransport::USBPRINTTransport(const std::string& path)
    : path_(path) {}

USBPRINTTransport::~USBPRINTTransport() {
    if (!closed_) {
        try { close(); } catch (...) {}
    }
}

#ifndef _WIN32

// ---- Non-Windows placeholder implementation ----

void USBPRINTTransport::open() {
    throw std::runtime_error("USBPRINT transport is only supported on Windows");
}

void USBPRINTTransport::close() {
    closed_ = true;
}

void USBPRINTTransport::write(const std::vector<unsigned char>&) {
    throw std::runtime_error("USBPRINT transport is only supported on Windows");
}

std::vector<unsigned char> USBPRINTTransport::read(int) {
    throw std::runtime_error("USBPRINT transport is only supported on Windows");
}

void USBPRINTTransport::drain() {
    throw std::runtime_error("USBPRINT transport is only supported on Windows");
}

std::string USBPRINTTransport::identify() {
    throw std::runtime_error("USBPRINT transport is only supported on Windows");
}

std::vector<std::string> enumeratePrinters() {
    throw std::runtime_error("printer enumeration is only supported on Windows");
}

#else

// ---- Windows implementation (Win32 USBPRINT + SetupDi) ----

#include <windows.h>
#include <setupapi.h>
#include <winioctl.h>

namespace {

const DWORD IOCTL_USBPRINT_GET_1284_ID = 2228276;
const DWORD IOCTL_USBPRINT_SOFT_RESET = 2228288;

const GUID GUID_DEVINTERFACE_USBPRINT = {
    0x28D78FAD, 0x5A12, 0x11D1,
    {0xAE, 0x5B, 0x00, 0x00, 0xF8, 0x03, 0xA8, 0xC2}};

} // namespace

void USBPRINTTransport::open() {
    std::wstring wpath(path_.begin(), path_.end());
    HANDLE h = CreateFileW(
        wpath.c_str(),
        GENERIC_READ | GENERIC_WRITE,
        FILE_SHARE_READ | FILE_SHARE_WRITE,
        nullptr,
        OPEN_EXISTING,
        FILE_FLAG_NO_BUFFERING | FILE_FLAG_WRITE_THROUGH,
        nullptr);
    if (h == INVALID_HANDLE_VALUE) {
        throw std::runtime_error("CreateFileW failed for " + path_);
    }
    handle_ = h;

    DWORD ret = 0;
    if (!DeviceIoControl(h, IOCTL_USBPRINT_SOFT_RESET, nullptr, 0, nullptr, 0, &ret, nullptr)) {
        CloseHandle(h);
        handle_ = nullptr;
        throw std::runtime_error("soft reset failed for " + path_);
    }
    closed_ = false;
}

void USBPRINTTransport::close() {
    if (handle_) {
        CloseHandle(static_cast<HANDLE>(handle_));
        handle_ = nullptr;
    }
    closed_ = true;
}

void USBPRINTTransport::write(const std::vector<unsigned char>& data) {
    if (closed_) throw std::runtime_error("USBPRINT device closed: " + path_);
    DWORD written = 0;
    if (!WriteFile(static_cast<HANDLE>(handle_), data.data(), static_cast<DWORD>(data.size()),
                   &written, nullptr)) {
        throw std::runtime_error("WriteFile failed");
    }
    if (written != data.size()) {
        throw std::runtime_error("wrote partial data to " + path_);
    }
}

std::vector<unsigned char> USBPRINTTransport::read(int size) {
    if (closed_) throw std::runtime_error("USBPRINT device closed: " + path_);
    while (static_cast<int>(buffer_.size()) < size) {
        std::vector<unsigned char> chunk(0x400000);
        DWORD read = 0;
        if (!ReadFile(static_cast<HANDLE>(handle_), chunk.data(),
                      static_cast<DWORD>(chunk.size()), &read, nullptr)) {
            throw std::runtime_error("ReadFile failed");
        }
        chunk.resize(read);
        buffer_.insert(buffer_.end(), chunk.begin(), chunk.end());
    }
    std::vector<unsigned char> out(buffer_.begin(), buffer_.begin() + size);
    buffer_.erase(buffer_.begin(), buffer_.begin() + size);
    return out;
}

void USBPRINTTransport::drain() {
    if (closed_) throw std::runtime_error("USBPRINT device closed: " + path_);
    for (;;) {
        std::vector<unsigned char> chunk(0x400000);
        DWORD read = 0;
        if (!ReadFile(static_cast<HANDLE>(handle_), chunk.data(),
                      static_cast<DWORD>(chunk.size()), &read, nullptr)) {
            throw std::runtime_error("ReadFile failed");
        }
        if (read == 0) return;
    }
}

std::string USBPRINTTransport::identify() {
    if (closed_) throw std::runtime_error("USBPRINT device closed: " + path_);
    std::vector<unsigned char> out(1024);
    DWORD ret = 0;
    if (!DeviceIoControl(static_cast<HANDLE>(handle_), IOCTL_USBPRINT_GET_1284_ID,
                         nullptr, 0, out.data(), static_cast<DWORD>(out.size()),
                         &ret, nullptr)) {
        throw std::runtime_error("identify failed");
    }
    out.resize(ret);
    // First two bytes are a little-endian length prefix.
    if (out.size() < 2) return "";
    return std::string(out.begin() + 2, out.end());
}

std::vector<std::string> enumeratePrinters() {
    std::vector<std::string> paths;
    HDEVINFO devInfo = SetupDiGetClassDevsW(
        &GUID_DEVINTERFACE_USBPRINT, nullptr, nullptr,
        DIGCF_PRESENT | DIGCF_DEVICEINTERFACE);
    if (devInfo == INVALID_HANDLE_VALUE) {
        throw std::runtime_error("SetupDiGetClassDevs failed");
    }

    SP_DEVICE_INTERFACE_DATA ifaceData = {};
    ifaceData.cbSize = sizeof(ifaceData);
    DWORD idx = 0;
    while (SetupDiEnumDeviceInterfaces(devInfo, nullptr, &GUID_DEVINTERFACE_USBPRINT,
                                       idx, &ifaceData)) {
        DWORD required = 0;
        SetupDiGetDeviceInterfaceDetailW(devInfo, &ifaceData, nullptr, 0, &required, nullptr);

        std::vector<unsigned char> buf(required);
        auto* detail = reinterpret_cast<PSP_DEVICE_INTERFACE_DETAIL_DATA_W>(buf.data());
        detail->cbSize = sizeof(SP_DEVICE_INTERFACE_DETAIL_DATA_W);

        if (SetupDiGetDeviceInterfaceDetailW(devInfo, &ifaceData, detail, required,
                                             nullptr, nullptr)) {
            std::wstring wpath(detail->DevicePath);
            std::string spath(wpath.begin(), wpath.end());
            paths.push_back(spath);
        }
        idx++;
    }
    SetupDiDestroyDeviceInfoList(devInfo);
    return paths;
}

#endif

} // namespace ezreset
