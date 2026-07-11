// Package app contains the Wails application logic: it exposes bindable
// methods that the frontend calls to list printers, read status and reset
// waste counters. It wraps the existing ezreset protocol packages.
// The desktop application is branded "Bustamante Print Tools".
package app

import (
	"context"
	_ "embed"
	"fmt"
	"sort"
	"strings"

	"ezreset/internal/devices"
	"ezreset/internal/d4"
	"ezreset/internal/printer"
	"ezreset/internal/transport"
)

//go:embed devices.xml
var devicesXML []byte

// PrinterInfo is the bindable representation of a connected printer.
type PrinterInfo struct {
	Path   string `json:"path"`
	Model  string `json:"model"`
	Des    string `json:"des"`
	Mfg    string `json:"mfg"`
	Serial string `json:"serial"`
}

// InkLevel is the bindable representation of an ink cartridge level.
type InkLevel struct {
	Color  string `json:"color"`
	Level  int    `json:"level"`
	Status string `json:"status"`
}

// WasteCounter is the bindable representation of a waste ink counter.
type WasteCounter struct {
	Index int     `json:"index"`
	Value int     `json:"value"`
	Max   int     `json:"max"`
	Ratio float64 `json:"ratio"`
}

// StatusView is the bindable representation of the full printer status.
type StatusView struct {
	State         string         `json:"state"`
	Error         string         `json:"error"`
	Source        string         `json:"source"`
	Serial        string         `json:"serial"`
	InkLevels     []InkLevel     `json:"inkLevels"`
	WasteCounters []WasteCounter `json:"wasteCounters"`
}

// App is the Wails application struct.
type App struct {
	ctx     context.Context
	db      map[string]devices.Device
	dbError error
}

// New creates the application, loading the device database.
func New() *App {
	db, err := loadDB()
	if err != nil {
		return &App{dbError: err}
	}
	return &App{db: db}
}

// Startup is called by Wails when the app starts.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// DBError returns the startup error, if any.
func (a *App) DBError() string {
	if a.dbError != nil {
		return a.dbError.Error()
	}
	return ""
}

// ListPrinters returns the connected USBPRINT printers, enriched with their
// IEEE 1284.4 identity (model, manufacturer, description, serial).
func (a *App) ListPrinters() ([]PrinterInfo, error) {
	paths, err := transport.EnumeratePrinters()
	if err != nil {
		return nil, err
	}

	var out []PrinterInfo
	for _, p := range paths {
		info := PrinterInfo{Path: p}
		// Best-effort identification: open the transport just long enough to
		// read the device ID. Failures are non-fatal (the printer is still listed).
		if t := transport.NewUSBPRINTTransport(p); t != nil {
			if openErr := t.Open(); openErr == nil {
				if id, idErr := t.Identify(); idErr == nil {
					info.Model = parseField(id, "MDL")
					info.Des = parseField(id, "DES")
					info.Mfg = parseField(id, "MFG")
					info.Serial = parseField(id, "SN")
				}
				_ = t.Close()
			}
		}
		out = append(out, info)
	}
	return out, nil
}

// GetStatus opens the printer at the given path and returns its status.
func (a *App) GetStatus(path string) (*StatusView, error) {
	p, err := a.openPrinter(path)
	if err != nil {
		return nil, err
	}
	defer p.Close()

	st, err := p.Printer.GetStatus()
	if err != nil {
		return nil, err
	}

	view := &StatusView{
		State:         st.State.String(),
		Error:         st.Error.String(),
		Source:        st.Source.String(),
		Serial:        st.Serial,
		InkLevels:     []InkLevel{},
		WasteCounters: []WasteCounter{},
	}

	for _, lvl := range st.Levels {
		view.InkLevels = append(view.InkLevels, InkLevel{
			Color:  lvl.Color.String(),
			Level:  lvl.Level,
			Status: lvl.Status.String(),
		})
	}

	wastes, err := p.Printer.GetWaste()
	if err != nil {
		return nil, err
	}
	for i, w := range wastes {
		ratio := 0.0
		if w[1] > 0 {
			ratio = float64(w[0]) / float64(w[1])
		}
		view.WasteCounters = append(view.WasteCounters, WasteCounter{
			Index: i,
			Value: w[0],
			Max:   w[1],
			Ratio: ratio,
		})
	}

	return view, nil
}

// ResetWaste opens the printer and resets all waste counters.
func (a *App) ResetWaste(path string) (string, error) {
	p, err := a.openPrinter(path)
	if err != nil {
		return "", err
	}
	defer p.Close()

	if err := p.Printer.ResetWaste(); err != nil {
		return "", err
	}
	return "Waste ink counters have been reset. You must now restart the printer.", nil
}

// Models returns the list of supported printer models.
func (a *App) Models() ([]string, error) {
	if a.db == nil {
		return nil, fmt.Errorf("device database not loaded")
	}
	models := make([]string, 0, len(a.db))
	for m := range a.db {
		models = append(models, m)
	}
	sort.Strings(models)
	return models, nil
}

// openedPrinter bundles a transport/backend with the high-level Printer.
type openedPrinter struct {
	Transport transport.Transport
	Backend   *d4.ControlBackend
	Printer   *printer.Printer
}

func (o *openedPrinter) Close() {
	if o.Backend != nil {
		_ = o.Backend.Close()
	}
	if o.Transport != nil {
		_ = o.Transport.Close()
	}
}

func (a *App) openPrinter(path string) (*openedPrinter, error) {
	t := transport.NewUSBPRINTTransport(path)
	if err := t.Open(); err != nil {
		return nil, fmt.Errorf("failed to open transport: %w", err)
	}

	backend := d4.NewControlBackend(t)
	if err := backend.Open(); err != nil {
		_ = t.Close()
		return nil, fmt.Errorf("failed to open control channel: %w", err)
	}

	id, err := backend.Identify()
	if err != nil {
		_ = backend.Close()
		return nil, fmt.Errorf("failed to identify printer: %w", err)
	}

	model := parseField(id, "MDL")
	dev, err := devices.ByModel(a.db, model)
	if err != nil {
		_ = backend.Close()
		return nil, err
	}

	return &openedPrinter{
		Transport: t,
		Backend:   backend,
		Printer:   printer.New(backend, dev),
	}, nil
}

func parseField(id, key string) string {
	for _, field := range strings.Split(id, ";") {
		if len(field) > len(key)+1 && strings.HasPrefix(field, key+":") {
			return field[len(key)+1:]
		}
	}
	return ""
}

func loadDB() (map[string]devices.Device, error) {
	// Prefer the embedded copy so the app works regardless of working directory.
	if db, err := devices.LoadFromBytes(devicesXML); err == nil {
		return db, nil
	}
	candidates := []string{
		"devices.xml",
		"internal/devices/devices.xml",
		"src/ez_reset/devices.xml",
		"frontend/dist/devices.xml",
	}
	for _, c := range candidates {
		if db, err := devices.Load(c); err == nil {
			return db, nil
		}
	}
	return nil, fmt.Errorf("devices.xml not found (looked in %v)", candidates)
}
