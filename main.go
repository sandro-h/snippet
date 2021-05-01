package main

import (
	"fmt"

	"fyne.io/fyne/app"
	"fyne.io/fyne/container"
	"fyne.io/fyne/widget"
	"github.com/go-vgo/robotgo"
	// "github.com/go-vgo/robotgo"
)

func main() {
	fmt.Println("hi!")
	// robotgo.TypeStr("Hello World")
	a := app.New()
	w := a.NewWindow("Hello")

	hello := widget.NewLabel("Hello Fyne!")
	w.SetContent(container.NewVBox(
		hello,
		widget.NewButton("Hi!", func() {
			robotgo.MoveMouse(100, 100)
		}),
	))

	w.ShowAndRun()
}
