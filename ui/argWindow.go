package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ArgWindow provides a set of input boxes to fill out arguments of a snippet.
type ArgWindow struct {
	win          fyne.Window
	entries      []*typeableEntry
	focusedIndex int
}

// NewArgWindow creates a new ArgWindow.
func NewArgWindow(win fyne.Window) *ArgWindow {
	return &ArgWindow{
		win: win,
	}
}

// ShowWithArgs shows the ArgWindow with the given list of arguments.
func (w *ArgWindow) ShowWithArgs(args []string, onSubmit func(map[string]string), onCancel func()) {
	w.entries = make([]*typeableEntry, 0)
	cont := container.NewVBox()
	for _, a := range args {
		lbl := widget.NewLabel(a)

		entry := newTypeableEntry()
		entry.onTypedKey = func(key *fyne.KeyEvent) {
			if key.Name == "Down" {
				w.focusEntry((w.focusedIndex + 1) % len(w.entries))
			} else if key.Name == "Up" {
				w.focusEntry((len(w.entries) + w.focusedIndex - 1) % len(w.entries))
			} else if key.Name == "Return" {
				vals := make(map[string]string)
				for i, e := range w.entries {
					vals[args[i]] = e.Text
				}
				w.win.Hide()
				onSubmit(vals)
			} else if key.Name == "Escape" {
				w.win.Hide()
				onCancel()
			}
		}

		cont.Add(lbl)
		cont.Add(entry)
		w.entries = append(w.entries, entry)
	}
	w.win.SetContent(cont)
	w.win.Resize(fyne.NewSize(300, cont.Size().Height+20))
	w.win.CenterOnScreen()
	w.focusEntry(0)
	w.win.Show()
}

func (w *ArgWindow) focusEntry(index int) {
	w.focusedIndex = index
	w.win.Canvas().Focus(w.entries[index])
}
