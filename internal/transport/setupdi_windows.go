//go:build windows

package transport

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// SPDeviceInterfaceData mirrors the Win32 SP_DEVICE_INTERFACE_DATA struct.
type SPDeviceInterfaceData struct {
	CbSize             uint32
	InterfaceClassGuid windows.GUID
	Flags              uint32
	Reserved           uintptr
}

// SPDeviceInterfaceDetailData mirrors the Win32 SP_DEVICE_INTERFACE_DETAIL_DATA struct.
type SPDeviceInterfaceDetailData struct {
	CbSize     uint32
	DevicePath [1]uint16
}

func setupDiGetClassDevs(guid *windows.GUID, flags uint32) (windows.Handle, error) {
	ret, _, err := procSetupDiGetClassDevs.Call(
		uintptr(unsafe.Pointer(guid)),
		0,
		0,
		uintptr(flags),
	)
	if ret == 0 {
		return 0, fmt.Errorf("SetupDiGetClassDevs: %w", err)
	}
	return windows.Handle(ret), nil
}

func setupDiEnumDeviceInterfaces(devInfo windows.Handle, guid *windows.GUID, index uint32, ifaceData *SPDeviceInterfaceData) (bool, error) {
	ret, _, err := procSetupDiEnumDeviceInterfaces.Call(
		uintptr(devInfo),
		0,
		uintptr(unsafe.Pointer(guid)),
		uintptr(index),
		uintptr(unsafe.Pointer(ifaceData)),
	)
	if ret == 0 {
		if errno, ok := err.(windows.Errno); ok && int(errno) == errorNoMoreItems {
			return false, nil
		}
		return false, fmt.Errorf("SetupDiEnumDeviceInterfaces: %w", err)
	}
	return true, nil
}

func setupDiGetDeviceInterfaceDetailSize(devInfo windows.Handle, ifaceData *SPDeviceInterfaceData, required *uint32) error {
	ret, _, err := procSetupDiGetDeviceInterfaceDetail.Call(
		uintptr(devInfo),
		uintptr(unsafe.Pointer(ifaceData)),
		0,
		0,
		uintptr(unsafe.Pointer(required)),
		0,
	)
	if ret == 0 {
		// ERROR_INSUFFICIENT_BUFFER is expected here.
		if errno, ok := err.(windows.Errno); ok && int(errno) == errorInsufficientBuf {
			return nil
		}
		return fmt.Errorf("SetupDiGetDeviceInterfaceDetail (size): %w", err)
	}
	return nil
}

func setupDiGetDeviceInterfaceDetail(devInfo windows.Handle, ifaceData *SPDeviceInterfaceData, detail *byte, size uint32) error {
	ret, _, err := procSetupDiGetDeviceInterfaceDetail.Call(
		uintptr(devInfo),
		uintptr(unsafe.Pointer(ifaceData)),
		uintptr(unsafe.Pointer(detail)),
		uintptr(size),
		0,
		0,
	)
	if ret == 0 {
		return fmt.Errorf("SetupDiGetDeviceInterfaceDetail: %w", err)
	}
	return nil
}

// devicePathToString reads the UTF-16 device path from a detail buffer whose
// first 4 bytes are CbSize and whose path begins immediately after.
func devicePathToString(buf []byte) string {
	hdr := (*[2]uint16)(unsafe.Pointer(&buf[0]))
	return windows.UTF16ToString((*[1 << 16]uint16)(unsafe.Pointer(&hdr[1]))[:])
}

var (
	modSetupapi                         = windows.NewLazySystemDLL("setupapi.dll")
	procSetupDiGetClassDevs             = modSetupapi.NewProc("SetupDiGetClassDevsW")
	procSetupDiEnumDeviceInterfaces     = modSetupapi.NewProc("SetupDiEnumDeviceInterfaces")
	procSetupDiGetDeviceInterfaceDetail = modSetupapi.NewProc("SetupDiGetDeviceInterfaceDetailW")
)
