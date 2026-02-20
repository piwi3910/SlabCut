package ui

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"github.com/piwi3910/SlabCut/internal/assets"
	"github.com/piwi3910/SlabCut/internal/version"
)

// ShowSplash displays an undecorated splash screen with the app logo and version.
// After the given duration it closes the splash and calls onDone.
func ShowSplash(app fyne.App, duration time.Duration, onDone func()) {
	splash := app.NewWindow("SlabCut")
	splash.SetFixedSize(true)
	splash.CenterOnScreen()

	img := canvas.NewImageFromResource(fyne.NewStaticResource("splash.png", assets.SplashPNG))
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(800, 600))

	versionText := canvas.NewText("SlabCut "+version.Short(), color.White)
	versionText.TextSize = 14
	versionText.Alignment = fyne.TextAlignCenter

	content := container.NewStack(
		img,
		container.NewBorder(nil, container.NewCenter(versionText), nil, nil),
	)

	splash.SetContent(content)
	splash.Resize(fyne.NewSize(800, 600))
	splash.SetPadded(false)
	splash.Show()

	go func() {
		time.Sleep(duration)
		splash.Close()
		if onDone != nil {
			onDone()
		}
	}()
}
