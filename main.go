package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
)

func main() {
	a := app.New()
	w := a.NewWindow("TLB decoder")
	w.Resize(fyne.NewSize(1024, 768))
	a.Settings().SetTheme(theme.DarkTheme())
	w.SetContent(tlbDecodingWindow())

	w.ShowAndRun()
}
