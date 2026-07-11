// Package utils contains helpers for parsing binary status structures and
// IEEE 1284.4 device identifier strings, ported from the Python ez-reset tool.
package utils

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// StatusEntry is a single entry inside a binary status struct.
type StatusEntry struct {
	Header  byte
	Length  byte
	Payload []byte
}

// ParseStatusStruct iterates over a binary status struct's entries.
//
// The status object contains various fields of interest, which are comprised of:
//   - header  - 1 byte
//   - size    - 1 byte
//   - payload - n bytes
func ParseStatusStruct(data []byte) ([]StatusEntry, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("status payload too short")
	}

	length := binary.LittleEndian.Uint16(data[0:2])
	if len(data) != int(length)+2 {
		return nil, fmt.Errorf("status payload length invalid: got %d, expected %d", len(data), int(length)+2)
	}

	var entries []StatusEntry

	index := 2
	for index < int(length) {
		header := data[index]
		index++

		parameterLength := int(data[index])
		index++

		if index+parameterLength > len(data) {
			return nil, fmt.Errorf("status entry truncated")
		}

		payload := data[index : index+parameterLength]
		index += parameterLength

		entries = append(entries, StatusEntry{
			Header:  header,
			Length:  byte(parameterLength),
			Payload: payload,
		})
	}

	return entries, nil
}

// ParseIdentifier parses an IEEE 1284.4 ID string into a map of key/value pairs.
func ParseIdentifier(identifier string) map[string]string {
	result := make(map[string]string)

	for _, field := range strings.Split(identifier, ";") {
		if field == "" {
			continue
		}

		parts := strings.SplitN(field, ":", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}

	return result
}
