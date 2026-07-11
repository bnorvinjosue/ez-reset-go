// Command bustamante-print-tools is a Go port of the ez-reset tool, with a
// Wails-based GUI branded "Bustamante Print Tools".
//
// On Windows it can talk to Epson printers over USB. On other platforms the
// USB transport is unavailable, but the GUI and device database still work.
package main

import (
	"embed"

	"ezreset/internal/app"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	application := app.New()

	err := wails.Run(&options.App{
		Title:  "Bustamante Print Tools",
		Width:  920,
		Height: 680,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 17, G: 24, B: 39, A: 255},
		OnStartup:         application.Startup,
		Bind: []interface{}{
			application,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
