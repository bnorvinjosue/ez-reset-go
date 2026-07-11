# ez-reset (Go port)

A Go port of [ez-reset](https://github.com/CiRIP/ez-reset): a tool to reset waste
ink counters on Epson printers over USB.

It implements the same protocols and logic as the original Python/Tkinter app:

- **D4** (IEEE 1284.4 / Dot4) control backend
- **END4** (Epson proprietary) control backend
- USBPRINT transport on Windows (`CreateFileW` / `DeviceIoControl`)
- Printer enumeration via the Win32 `SetupDi*` API
- Status parsing, EEPROM read/write, and waste counter reset
- Device database loaded from `devices.xml`

## GUI (Wails)

The main interface is a [Wails](https://wails.io) desktop app with a modern
dark-themed UI:

- Scans for connected USB printers
- Shows ink levels as colored gauges
- Shows waste ink counters as progress bars (turn red when near full)
- One-click "Reset all waste counters" button
- Lists all supported printer models

### Building the GUI

```sh
# Install the Wails CLI (once)
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Linux (needs webkit2gtk-4.1 dev packages)
sudo apt-get install libwebkit2gtk-4.1-dev libgtk-3-dev
wails build -tags webkit2_41

# Windows (full USB functionality)
wails build

# macOS
wails build
```

The built binary is in `build/bin/ezreset`. On Windows it can talk to real
Epson printers over USB; on other platforms the GUI and device database work
but the USB transport is unavailable.

### Building for Windows from Linux/macOS

Cross-compilation is fully supported (no Windows machine needed):

```sh
# One-shot script:
./build-windows.sh

# Or manually:
( cd frontend && npm install && npm run build )
wails build -platform windows/amd64
```

This produces `build/bin/ezreset.exe` — a self-contained `PE32+` GUI executable
that embeds the frontend and `devices.xml`. On the target Windows machine it
needs the **WebView2 runtime** (preinstalled on Windows 10/11) and the USBPRINT
driver (provided by the Epson USB driver). No extra files are required next to
the `.exe`.

### Running the dev server

```sh
wails dev
```

## CLI (legacy)

A small CLI is also available via the protocol packages. Build it with:

```sh
go build -o ezreset-cli .   # (the wails main.go replaces the old CLI entrypoint)
```

The protocol/device logic lives under `internal/` and is fully testable with
`go test ./...`.

## Platform notes

The USBPRINT transport and printer enumeration are guarded by Go build tags
(`//go:build windows`). On other platforms these return an error, but the rest
of the code (XML database, status parsing, protocol logic) compiles and can be
tested.

## License

Same as the original project. The `winapi`/SetupDi bindings are derived from
pywinusb (BSD-licensed), as in the original.
