package ui

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/sandro-h/snippet/util"
)

// SearchWidget provides fuzzy search for a list of snippets. The snippets matching the search are displayed in a navigable list.
type SearchWidget struct {
	snippets         []*util.Snippet
	snippetsText     []string
	filteredSnippets []*util.Snippet
	selectedID       widget.ListItemID
	onSubmit         func(snippet *util.Snippet)
	onCancel         func()
	List             *widget.List
	Entry            *typeableEntry
}

// NewSearchWidget creates a new SearchWidget.
func NewSearchWidget(snippets []*util.Snippet, onSubmit func(snippet *util.Snippet), onCancel func()) *SearchWidget {
	w := &SearchWidget{onSubmit: onSubmit, onCancel: onCancel}
	w.createList()
	w.createEntry()
	w.SetSnippets(snippets)
	return w
}

// SetSnippets sets a new list of snippets for the widget to display.
func (w *SearchWidget) SetSnippets(snippets []*util.Snippet) {
	var snippetsText []string
	for _, s := range snippets {
		snippetsText = append(snippetsText, fmt.Sprintf("%s: %s", s.Label, s.Content))
	}
	w.snippetsText = snippetsText
	w.snippets = snippets
	w.Entry.OnChanged(w.Entry.Text)
}

func (w *SearchWidget) GetSnippets() []*util.Snippet {
	return w.snippets
}

func (w *SearchWidget) createList() {
	w.List = widget.NewList(
		func() int {
			return len(w.filteredSnippets)
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("tmpl lbl")
			label.TextStyle.Bold = true
			content := canvas.NewText("tmpl content", color.RGBA{128, 128, 128, 255})
			content.TextStyle.Monospace = true
			return container.NewHBox(label, content)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			container := item.(*fyne.Container)
			label := container.Objects[0].(*widget.Label)
			content := container.Objects[1].(*canvas.Text)
			label.SetText(w.filteredSnippets[id].Label)
			content.Text = strings.ReplaceAll(w.filteredSnippets[id].Content, "\n", "\\n")
			ellipsis(container, content)
		},
	)

	w.List.OnSelected = func(id widget.ListItemID) {
		w.selectedID = id
	}
	w.List.Select(0)
}

func (w *SearchWidget) createEntry() {
	w.Entry = newTypeableEntry()

	resetSearch := func() {
		w.Entry.Text = ""
		w.Entry.OnChanged(w.Entry.Text)
		w.List.Select(0)
	}

	w.Entry.onTypedKey = func(key *fyne.KeyEvent) {
		if key.Name == "Down" {
			w.List.Select((w.selectedID + 1) % len(w.filteredSnippets))
		} else if key.Name == "Up" {
			w.List.Select((len(w.filteredSnippets) + w.selectedID - 1) % len(w.filteredSnippets))
		} else if key.Name == "Return" {
			if w.selectedID >= 0 && w.selectedID < len(w.filteredSnippets) {
				w.onSubmit(w.filteredSnippets[w.selectedID])
				resetSearch()
			}
		} else if key.Name == "Escape" {
			w.onCancel()
			resetSearch()
		}
	}
	w.Entry.OnChanged = func(s string) {
		matches := util.SearchFuzzy(s, w.snippetsText)
		w.filteredSnippets = make([]*util.Snippet, 0)
		for _, m := range matches {
			w.filteredSnippets = append(w.filteredSnippets, w.snippets[m.Index])
		}
		w.List.Refresh()
		w.List.Select(0)
	}
}

type typeableEntry struct {
	widget.Entry
	onTypedKey func(key *fyne.KeyEvent)
}

func newTypeableEntry() *typeableEntry {
	e := &typeableEntry{}
	e.ExtendBaseWidget(e)
	return e
}

func (e *typeableEntry) TypedKey(key *fyne.KeyEvent) {
	e.Entry.TypedKey(key)
	if e.onTypedKey != nil {
		e.onTypedKey(key)
	}
}

func ellipsis(container *fyne.Container, label *canvas.Text) {
	w := fyne.MeasureText(string(label.Text), theme.TextSize(), label.TextStyle).Width
	if label.Position().X+w > container.Size().Width {
		wellipsis := fyne.MeasureText(string("..."), theme.TextSize(), label.TextStyle).Width
		wmax := container.Size().Width - wellipsis
		wpc := float64(w) / float64(len(label.Text))
		k := 0
		for label.Position().X+w > wmax {
			overlap := label.Position().X + w - wmax
			overlapc := int(math.Ceil(math.Max(1, float64(overlap)/wpc)))
			if overlapc > len(label.Text) {
				break
			}
			label.Text = label.Text[:len(label.Text)-overlapc]
			w = fyne.MeasureText(string(label.Text), theme.TextSize(), label.TextStyle).Width
			k++
		}
		label.Text += "..."
		label.Refresh()
	}
}
