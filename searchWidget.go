package main

import (
	"fmt"
	"image/color"
	"math"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type searchWidget struct {
	snippets         []*snippet
	snippetsText     []string
	filteredSnippets []*snippet
	selectedID       widget.ListItemID
	onSubmit         func(snippet *snippet)
	onCancel         func()
	list             *widget.List
	entry            *searchEntry
}

func newSearchWidget(snippets []*snippet, onSubmit func(snippet *snippet), onCancel func()) *searchWidget {
	w := &searchWidget{onSubmit: onSubmit, onCancel: onCancel}
	w.createList()
	w.createEntry()
	w.setSnippets(snippets)
	return w
}

func (w *searchWidget) setSnippets(snippets []*snippet) {
	var snippetsText []string
	for _, s := range snippets {
		snippetsText = append(snippetsText, fmt.Sprintf("%s: %s", s.label, s.content))
	}
	w.snippetsText = snippetsText
	w.snippets = snippets
	w.entry.OnChanged(w.entry.Text)
}

func (w *searchWidget) createList() {
	w.list = widget.NewList(
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
			label.SetText(w.filteredSnippets[id].label)
			content.Text = strings.ReplaceAll(w.filteredSnippets[id].content, "\n", "\\n")
			ellipsis(container, content)
		},
	)

	w.list.OnSelected = func(id widget.ListItemID) {
		w.selectedID = id
	}
	w.list.Select(0)
}

func (w *searchWidget) createEntry() {
	w.entry = newSearchEntry()

	resetSearch := func() {
		w.entry.Text = ""
		w.entry.OnChanged(w.entry.Text)
		w.list.Select(0)
	}

	w.entry.onTypedKey = func(key *fyne.KeyEvent) {
		if key.Name == "Down" {
			w.list.Select((w.selectedID + 1) % len(w.filteredSnippets))
		} else if key.Name == "Up" {
			w.list.Select((len(w.filteredSnippets) + w.selectedID - 1) % len(w.filteredSnippets))
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
	w.entry.OnChanged = func(s string) {
		ranked := fuzzy.RankFindFold(s, w.snippetsText)
		sort.Sort(ranked)
		w.filteredSnippets = make([]*snippet, 0)
		for _, r := range ranked {
			w.filteredSnippets = append(w.filteredSnippets, w.snippets[r.OriginalIndex])
		}
		w.list.Refresh()
		w.list.Select(0)
	}
}

type searchEntry struct {
	widget.Entry
	onTypedKey func(key *fyne.KeyEvent)
}

func newSearchEntry() *searchEntry {
	e := &searchEntry{}
	e.ExtendBaseWidget(e)
	return e
}

func (e *searchEntry) TypedKey(key *fyne.KeyEvent) {
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