package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ShowPasswordWindow shows a window for the user to input their password.
func ShowPasswordWindow(w fyne.Window, label string, onSubmit func(string), onCancel func()) {
	lbl := widget.NewLabel(label)
	entry := newTypeablePasswordEntry()
	entry.onTypedKey = func(key *fyne.KeyEvent) {
		if key.Name == "Return" {
			w.Hide()
			onSubmit(entry.Text)
		} else if key.Name == "Escape" {
			w.Hide()
			onCancel()
		}
	}

	w.SetContent(container.NewVBox(lbl, entry))
	w.Resize(fyne.NewSize(300, 80))
	w.CenterOnScreen()
	w.Canvas().Focus(entry)
	w.Show()
}
