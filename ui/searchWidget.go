package ui

import (
	"math"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/sandro-h/snippet/fuzzy"
	"github.com/sandro-h/snippet/util"
)

type snippetSegment struct {
	str         string
	highlighted bool
}

type filteredSnippet struct {
	snippet            *util.Snippet
	highlightedLabel   []snippetSegment
	highlightedContent []snippetSegment
}

// SearchWidget provides fuzzy search for a list of snippets. The snippets matching the search are displayed in a navigable list.
type SearchWidget struct {
	snippets                []*util.Snippet
	snippetLabels           []string
	snippetContents         []string
	filteredSnippets        []*filteredSnippet
	selectedID              widget.ListItemID
	onSubmit                func(snippet *util.Snippet)
	onCancel                func()
	List                    *widget.List
	Entry                   *typeableEntry
	renderLock              sync.Mutex
	labelStyle              widget.RichTextStyle
	highlightedLabelStyle   widget.RichTextStyle
	contentStyle            widget.RichTextStyle
	highlightedContentStyle widget.RichTextStyle
}

// NewSearchWidget creates a new SearchWidget.
func NewSearchWidget(snippets []*util.Snippet, onSubmit func(snippet *util.Snippet), onCancel func()) *SearchWidget {
	w := &SearchWidget{
		onSubmit: onSubmit,
		onCancel: onCancel,
		labelStyle: widget.RichTextStyle{
			Inline:    true,
			TextStyle: fyne.TextStyle{Bold: true},
		},
		highlightedLabelStyle: widget.RichTextStyle{
			Inline:    true,
			TextStyle: fyne.TextStyle{Bold: true},
			ColorName: theme.ColorNamePrimary,
		},
		contentStyle: widget.RichTextStyle{
			ColorName: ColorSnippetContent,
			Inline:    true,
		},
		highlightedContentStyle: widget.RichTextStyle{
			ColorName: theme.ColorNamePrimary,
			Inline:    true,
		}}
	w.createList()
	w.createEntry()
	w.SetSnippets(snippets)
	return w
}

// SetSnippets sets a new list of snippets for the widget to display.
func (w *SearchWidget) SetSnippets(snippets []*util.Snippet) {
	var snippetLabels []string
	var snippetContents []string
	for _, s := range snippets {
		snippetLabels = append(snippetLabels, s.Label)
		snippetContents = append(snippetContents, strings.ReplaceAll(s.Content, "\n", "\\n"))
	}
	w.snippetLabels = snippetLabels
	w.snippetContents = snippetContents
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
			if id >= len(w.filteredSnippets) {
				return
			}

			container := item.(*fyne.Container)
			label := container.Objects[0].(*widget.RichText)
			content := container.Objects[1].(*widget.RichText)

			w.renderLock.Lock()
			label.Segments = createTextSegments(w.filteredSnippets[id].highlightedLabel, w.labelStyle, w.highlightedLabelStyle)
			content.Segments = createTextSegments(w.filteredSnippets[id].highlightedContent, w.contentStyle, w.highlightedContentStyle)
			w.renderLock.Unlock()

			ellipsis(container, content, w.contentStyle)
			label.Refresh()
			content.Refresh()
		},
	)

	w.List.OnSelected = func(id widget.ListItemID) {
		w.selectedID = id
	}
	w.List.Select(0)
}

func createTextSegments(segments []snippetSegment, style widget.RichTextStyle, highlightedStyle widget.RichTextStyle) []widget.RichTextSegment {
	var textSegments []widget.RichTextSegment
	for _, s := range segments {
		textSegment := widget.TextSegment{Text: s.str}
		if s.highlighted {
			textSegment.Style = highlightedStyle
		} else {
			textSegment.Style = style
		}

		textSegments = append(textSegments, &textSegment)
	}
	return textSegments
}

func (w *SearchWidget) createEntry() {
	w.Entry = newTypeableEntry()

	resetSearch := func(retainSelection bool) {
		selectedLabel := w.filteredSnippets[w.selectedID].snippet.Label
		w.Entry.Text = ""
		w.Entry.OnChanged(w.Entry.Text)

		if retainSelection {
			newIndex := -1
			for i, s := range w.filteredSnippets {
				if s.snippet.Label == selectedLabel {
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
				w.onSubmit(w.filteredSnippets[w.selectedID].snippet)
				resetSearch(true)
			}
		} else if key.Name == "Escape" {
			w.onCancel()
			resetSearch(false)
		}
	}
	w.Entry.OnChanged = func(s string) {
		matches := fuzzy.SearchFuzzyMulti(s, w.snippetLabels, w.snippetContents)
		var filteredSnippets []*filteredSnippet

		for _, m := range matches {
			highlightedLabel := createHighlightedSegments(w.snippetLabels[m.Index], m.Match1)
			highlightedContent := createHighlightedSegments(w.snippetContents[m.Index], m.Match2)

			s := &filteredSnippet{
				snippet:            w.snippets[m.Index],
				highlightedLabel:   highlightedLabel,
				highlightedContent: highlightedContent,
			}
			filteredSnippets = append(filteredSnippets, s)
		}

		w.renderLock.Lock()
		w.filteredSnippets = filteredSnippets
		w.renderLock.Unlock()

		w.List.Refresh()
		w.List.Select(0)
	}
}

func createHighlightedSegments(text string, match fuzzy.Match) []snippetSegment {
	segments := []snippetSegment{{str: text, highlighted: false}}
	if match.Index > -1 {
		segments = make([]snippetSegment, 0)

		li := 0
		for _, r := range match.MatchedRanges {
			if r.Start >= len(text) {
				break
			}
			if r.Start > li {
				segments = append(segments, snippetSegment{str: text[li:r.Start], highlighted: false})
			}
			segments = append(segments, snippetSegment{str: text[r.Start : r.End+1], highlighted: true})
			li = r.End + 1
		}
		if li < len(text) {
			segments = append(segments, snippetSegment{str: text[li:], highlighted: false})
		}
	}
	return segments
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
		wmax := getMaxContentWidth(container, label) - wellipsis
		wpc := float64(w) / float64(len(label.String()))
		startEllipsis := false

		for w > wmax {
			overlap := w - wmax
			overlapc := int(math.Ceil(math.Max(1, float64(overlap)/wpc)))
			if overlapc > len(label.String()) {
				break
			}

			truncatedFirst := truncateSegments(&label.Segments, overlapc)
			if !startEllipsis && truncatedFirst {
				startEllipsis = true
				wmax -= wellipsis
			}

			label.Refresh()
			w = measureRichText(label)
		}

		if startEllipsis {
			firstSeg := label.Segments[0].(*widget.TextSegment)
			firstSeg.Text = "..." + firstSeg.Text
		}
		label.Segments = append(label.Segments, &widget.TextSegment{Text: "...", Style: ellipsisStyle})
		label.Refresh()
	}
}

func truncateSegments(segments *[]widget.RichTextSegment, truncateLen int) bool {
	firstSeg := (*segments)[0].(*widget.TextSegment)
	firstHighlighted := firstSeg.Style.ColorName == theme.ColorNamePrimary
	truncatedFirst := false

	for truncateLen > 0 {
		// If there's only two segments left and the first one is not highlighted,
		// remove content from the first one, to ensure the highlight is visible.
		if len(*segments) == 2 && !firstHighlighted && firstSeg.Text != "" {
			truncatedFirst = true

			if len(firstSeg.Textual()) <= truncateLen {
				truncateLen -= len(firstSeg.Textual())
				firstSeg.Text = ""
			} else {
				firstSeg.Text = firstSeg.Text[truncateLen:]
				truncateLen = 0
			}
		} else {
			lastSeg := (*segments)[len(*segments)-1].(*widget.TextSegment)

			if len(lastSeg.Textual()) <= truncateLen {
				*segments = (*segments)[:len(*segments)-1]
				truncateLen -= len(lastSeg.Textual())
			} else {
				lastSeg.Text = lastSeg.Text[:len(lastSeg.Text)-truncateLen]
				truncateLen = 0
			}
		}
	}

	return truncatedFirst
}

func getMaxContentWidth(container *fyne.Container, label *widget.RichText) float32 {
	return container.Size().Width - label.Position().X
}

func measureRichText(label *widget.RichText) float32 {
	var w float32 = 0.0
	for _, s := range label.Segments {
		ts := s.(*widget.TextSegment)
		w = w + fyne.MeasureText(s.Textual(), theme.TextSize(), ts.Style.TextStyle).Width
	}
	return w
}
