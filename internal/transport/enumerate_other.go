//go:build !windows

package transport

import "fmt"

// EnumeratePrinters is unsupported on non-Windows platforms.
func EnumeratePrinters() ([]string, error) {
	return nil, fmt.Errorf("printer enumeration is only supported on Windows")
}
