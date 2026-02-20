// SlabCut — CNC Cut List Optimizer with GCode Export
//
// A cross-platform desktop application for optimizing rectangular
// cut lists from stock sheets and exporting CNC-ready GCode.
//
// Build:
//   go build -o slabcut ./cmd/slabcut
//
// Cross-compile:
//   GOOS=windows GOARCH=amd64 go build -o slabcut.exe ./cmd/slabcut
//   GOOS=darwin  GOARCH=amd64 go build -o slabcut-darwin ./cmd/slabcut
//
// Using fyne-cross (recommended for proper packaging):
//   go install github.com/fyne-io/fyne-cross@latest
//   fyne-cross windows -arch=amd64
//   fyne-cross darwin  -arch=amd64,arm64

package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/piwi3910/SlabCut/internal/ui"
)

func main() {
	application := app.NewWithID("com.piwi3910.slabcut")
	window := application.NewWindow("SlabCut — CNC Cut List Optimizer")

	appUI := ui.NewApp(window)
	appUI.SetupMenus() // Setup the native menu bar
	window.SetContent(appUI.Build())
	window.Resize(fyne.NewSize(1000, 700))
	window.CenterOnScreen()
	window.ShowAndRun()
}
