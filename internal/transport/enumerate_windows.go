//go:build windows

package transport

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var guidDevinterfaceUsbprint = windows.GUID{
	Data1: 0x28D78FAD,
	Data2: 0x5A12,
	Data3: 0x11D1,
	Data4: [8]byte{0xAE, 0x5B, 0x00, 0x00, 0xF8, 0x03, 0xA8, 0xC2},
}

const (
	digcfPresent         = 2
	digcfDeviceinterface = 16
	errorNoMoreItems     = 259
	errorInsufficientBuf = 122
)

// EnumeratePrinters returns the device paths of all connected USBPRINT
// printers using the SetupDi* Win32 API, ported from the Python ez-reset tool.
func EnumeratePrinters() ([]string, error) {
	devInfo, err := setupDiGetClassDevs(&guidDevinterfaceUsbprint, digcfPresent|digcfDeviceinterface)
	if err != nil {
		return nil, err
	}
	defer windows.SetupDiDestroyDeviceInfoList(windows.DevInfo(devInfo))

	var paths []string
	idx := uint32(0)

	for {
		var ifaceData SPDeviceInterfaceData
		ifaceData.CbSize = uint32(unsafe.Sizeof(SPDeviceInterfaceData{}))

		ok, err := setupDiEnumDeviceInterfaces(devInfo, &guidDevinterfaceUsbprint, idx, &ifaceData)
		if !ok {
			if err == nil {
				break
			}
			return nil, err
		}

		// First call: get required size.
		required := uint32(0)
		if err := setupDiGetDeviceInterfaceDetailSize(devInfo, &ifaceData, &required); err != nil {
			return nil, err
		}

		if required < uint32(unsafe.Sizeof(SPDeviceInterfaceDetailData{})) {
			required = uint32(unsafe.Sizeof(SPDeviceInterfaceDetailData{}))
		}
		buf := make([]byte, required)
		detail := (*SPDeviceInterfaceDetailData)(unsafe.Pointer(&buf[0]))
		detail.CbSize = uint32(unsafe.Sizeof(SPDeviceInterfaceDetailData{}))

		if err := setupDiGetDeviceInterfaceDetail(devInfo, &ifaceData, &buf[0], required); err != nil {
			return nil, err
		}

		paths = append(paths, devicePathToString(buf))
		idx++
	}

	return paths, nil
}
