package main

import (
	"fmt"

	"fyne.io/fyne/app"
	"fyne.io/fyne/container"
	"fyne.io/fyne/widget"
	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

func main() {
	fmt.Println("hi!")
	// robotgo.TypeStr("Hello World")
	go add()
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

func add() {
	fmt.Println("--- Please press ctrl + shift + q to stop hook ---")
	robotgo.EventHook(hook.KeyDown, []string{"q", "ctrl", "shift"}, func(e hook.Event) {
		fmt.Println("ctrl-shift-q")
		robotgo.EventEnd()
	})

	fmt.Println("--- Please press w---")
	robotgo.EventHook(hook.KeyDown, []string{"w"}, func(e hook.Event) {
		fmt.Println("omg w was pressed")
	})

	s := robotgo.EventStart()
	<-robotgo.EventProcess(s)
}
