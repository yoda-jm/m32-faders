package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("M32 OSC Controller")

	hello := widget.NewLabel("Hello, Fyne!")
	myWindow.SetContent(container.NewVBox(
		hello,
	))

	myWindow.ShowAndRun()
}
