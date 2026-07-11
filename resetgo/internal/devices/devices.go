// Package devices loads printer definitions from devices.xml and resolves a
// Device for a given model string, ported from the Python ez-reset tool.
package devices

import (
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Counter describes a single waste ink counter: the EEPROM addresses that hold
// the current value and the maximum value.
type Counter struct {
	Addresses []int
	Max       int
}

// Device is the resolved definition for a printer model.
type Device struct {
	Model    []byte
	Key      []byte
	Counters []Counter
	Reset    map[int]int
}

// xmlRoot is the top-level <data> element.
type xmlRoot struct {
	XMLName  xml.Name `xml:"data"`
	Printers []struct {
		Title  string `xml:"title,attr"`
		Model  string `xml:"model,attr"`
		Specs  string `xml:"specs,attr"`
	} `xml:"records>printer"`
	Devices []xmlDevice `xml:"devices>device"`
}

// xmlDevice is a named <device> entry (e.g. <SC700>).
type xmlDevice struct {
	Name    string `xml:"-"`
	Service *struct {
		Factory *string `xml:"factory"`
		Keyword *string `xml:"keyword"`
	} `xml:"service"`
	Waste *struct {
		Query *struct {
			Counters []struct {
				Text string `xml:",chardata"`
				Max  string `xml:"max"`
			} `xml:"counter"`
		} `xml:"query"`
		Reset *string `xml:"reset"`
	} `xml:"waste"`
}

// UnmarshalXML captures the tag name of each <device> element.
func (d *xmlDevice) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	d.Name = start.Name.Local
	type alias xmlDevice
	aux := alias(*d)
	if err := dec.DecodeElement(&aux, &start); err != nil {
		return err
	}
	*d = xmlDevice(aux)
	return nil
}

func parseByteList(text string) ([]byte, error) {
	fields := strings.Fields(text)
	out := make([]byte, 0, len(fields))
	for _, f := range fields {
		v, err := strconv.ParseUint(f, 0, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid byte %q: %w", f, err)
		}
		out = append(out, byte(v))
	}
	return out, nil
}

func parseAddressList(text string) ([]int, error) {
	fields := strings.Fields(text)
	out := make([]int, 0, len(fields))
	for _, f := range fields {
		v, err := strconv.ParseInt(f, 0, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid address %q: %w", f, err)
		}
		out = append(out, int(v))
	}
	return out, nil
}

// Load parses devices.xml from the given path.
func Load(path string) (map[string]Device, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parse(raw)
}

// LoadFromBytes parses devices.xml from an in-memory byte slice.
func LoadFromBytes(raw []byte) (map[string]Device, error) {
	return parse(raw)
}

func parse(raw []byte) (map[string]Device, error) {
	var root xmlRoot
	if err := xml.Unmarshal(raw, &root); err != nil {
		return nil, err
	}

	// Index named device specs by tag name.
	specs := make(map[string]xmlDevice)
	for _, d := range root.Devices {
		specs[d.Name] = d
	}

	byModel := make(map[string]Device)
	for _, p := range root.Printers {
		if p.Model == "" || p.Title == "" || p.Model == "Device" {
			continue
		}
		dev := Device{Reset: map[int]int{}}

		for _, specName := range strings.Split(p.Specs, ",") {
			specName = strings.TrimSpace(specName)
			if specName == "" {
				continue
			}

			spec, ok := specs[specName]
			if !ok {
				continue
			}

			if spec.Service != nil && spec.Service.Factory != nil {
				model, err := parseByteList(*spec.Service.Factory)
				if err != nil {
					return nil, err
				}
				dev.Model = model

				if spec.Service.Keyword != nil {
					key, err := parseByteList(*spec.Service.Keyword)
					if err != nil {
						return nil, err
					}
					dev.Key = key
				}
			}

			if spec.Waste != nil {
				if spec.Waste.Query != nil {
					for _, c := range spec.Waste.Query.Counters {
						text := c.Text
						if strings.TrimSpace(text) == "" {
							continue
						}
						addresses, err := parseAddressList(text)
						if err != nil {
							return nil, err
						}
						maxVal := 0
						if c.Max != "" {
							maxVal, err = strconv.Atoi(strings.TrimSpace(c.Max))
							if err != nil {
								return nil, fmt.Errorf("invalid max %q: %w", c.Max, err)
							}
						}
						dev.Counters = append(dev.Counters, Counter{Addresses: addresses, Max: maxVal})
					}
				}

				if spec.Waste.Reset != nil {
					fields := strings.Fields(*spec.Waste.Reset)
					for i := 0; i+1 < len(fields); i += 2 {
						addr, err := strconv.ParseInt(fields[i], 0, 32)
						if err != nil {
							return nil, err
						}
						val, err := strconv.ParseInt(fields[i+1], 0, 32)
						if err != nil {
							return nil, err
						}
						dev.Reset[int(addr)] = int(val)
					}
				}
			}
		}

		byModel[p.Model] = dev
	}

	return byModel, nil
}

// ByModel returns the device definition for the given model string.
func ByModel(devices map[string]Device, model string) (Device, error) {
	dev, ok := devices[model]
	if !ok {
		return Device{}, fmt.Errorf("unknown printer model %q", model)
	}
	return dev, nil
}
