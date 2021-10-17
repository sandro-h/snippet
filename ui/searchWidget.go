package ui

import (
	"fmt"
	"math"
	"strings"

	"fyne.io/fyne/v2"
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

func (w *SearchWidget) createList() {
	w.List = widget.NewList(
		func() int {
			return len(w.filteredSnippets)
		},
		func() fyne.CanvasObject {
			label := widget.NewRichTextWithText("tmpl lbl")
			content := widget.NewRichTextWithText("tmpl content")
			return container.NewHBox(label, content)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			container := item.(*fyne.Container)
			label := container.Objects[0].(*widget.RichText)
			content := container.Objects[1].(*widget.RichText)

			labelStyle := widget.RichTextStyle{
				Inline:    true,
				TextStyle: fyne.TextStyle{Bold: true},
			}
			label.Segments = []widget.RichTextSegment{&widget.TextSegment{Style: labelStyle,
				Text: w.filteredSnippets[id].Label}}
			label.Refresh()

			text := strings.ReplaceAll(w.filteredSnippets[id].Content, "\n", "\\n")
			contentStyle := widget.RichTextStyle{
				ColorName: ColorSnippetContent,
				Inline:    true,
				TextStyle: fyne.TextStyle{
					// Not active because it is not aligned with label since upgrading to Fyne 2.1:
					// Monospace: true,
				},
			}

			content.Segments = []widget.RichTextSegment{&widget.TextSegment{Style: contentStyle,
				Text: text}}
			ellipsis(container, content, contentStyle)
			content.Refresh()
		},
	)

	w.List.OnSelected = func(id widget.ListItemID) {
		w.selectedID = id
	}
	w.List.Select(0)
}

func (w *SearchWidget) createEntry() {
	w.Entry = newTypeableEntry()

	resetSearch := func(retainSelection bool) {
		selectedLabel := w.filteredSnippets[w.selectedID].Label
		w.Entry.Text = ""
		w.Entry.OnChanged(w.Entry.Text)

		if retainSelection {
			newIndex := -1
			for i, s := range w.filteredSnippets {
				if s.Label == selectedLabel {
					newIndex = i
					break
				}
			}
			if newIndex > -1 {
				w.List.Select(newIndex)
			}
		} else {
			w.List.Select(0)
		}

	}

	w.Entry.onTypedKey = func(key *fyne.KeyEvent) {
		if key.Name == "Down" {
			w.List.Select((w.selectedID + 1) % len(w.filteredSnippets))
		} else if key.Name == "Up" {
			w.List.Select((len(w.filteredSnippets) + w.selectedID - 1) % len(w.filteredSnippets))
		} else if key.Name == "Return" {
			if w.selectedID >= 0 && w.selectedID < len(w.filteredSnippets) {
				w.onSubmit(w.filteredSnippets[w.selectedID])
				resetSearch(true)
			}
		} else if key.Name == "Escape" {
			w.onCancel()
			resetSearch(false)
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

func newTypeablePasswordEntry() *typeableEntry {
	e := &typeableEntry{
		widget.Entry{
			Password: true,
			Wrapping: fyne.TextTruncate,
		},
		nil,
	}
	e.ExtendBaseWidget(e)
	return e
}

func (e *typeableEntry) TypedKey(key *fyne.KeyEvent) {
	e.Entry.TypedKey(key)
	if e.onTypedKey != nil {
		e.onTypedKey(key)
	}
}

func ellipsis(container *fyne.Container, label *widget.RichText, ellipsisStyle widget.RichTextStyle) {
	w := measureRichText(label)
	if label.Position().X+w > container.Size().Width {
		wellipsis := fyne.MeasureText(string("..."), theme.TextSize(), ellipsisStyle.TextStyle).Width
		wmax := container.Size().Width - wellipsis
		wpc := float64(w) / float64(len(label.String()))
		k := 0

		for label.Position().X+w > wmax {
			overlap := label.Position().X + w - wmax
			overlapc := int(math.Ceil(math.Max(1, float64(overlap)/wpc)))
			if overlapc > len(label.String()) {
				break
			}

			tgtLen := len(label.String()) - overlapc
			for len(label.String()) > tgtLen {
				lastSeg := label.Segments[len(label.Segments)-1].(*widget.TextSegment)
				if len(lastSeg.Textual()) <= overlapc {
					label.Segments = label.Segments[:len(label.Segments)-1]
				} else {
					lastSeg.Text = lastSeg.Text[:len(lastSeg.Text)-overlapc]
				}
			}

			label.Refresh()
			w = measureRichText(label)
			k++
		}

		label.Segments = append(label.Segments, &widget.TextSegment{Text: "...", Style: ellipsisStyle})
		label.Refresh()
	}
}

func measureRichText(label *widget.RichText) float32 {
	var w float32 = 0.0
	for _, s := range label.Segments {
		ts := s.(*widget.TextSegment)
		w = w + fyne.MeasureText(s.Textual(), theme.TextSize(), ts.Style.TextStyle).Width
	}
	return w
}
