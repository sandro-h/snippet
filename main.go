package main

import (
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/go-vgo/robotgo"
	"github.com/lithammer/fuzzysearch/fuzzy"
	hook "github.com/robotn/gohook"
	"gopkg.in/yaml.v2"
)

type snippet struct {
	label   string
	content string
}

func main() {
	snippets, snippetsText, err := loadSnippets()
	if err != nil {
		panic(err)
	}

	a := app.New()
	a.Settings().SetTheme(&myTheme{})
	var w fyne.Window
	if drv, ok := fyne.CurrentApp().Driver().(desktop.Driver); ok {
		w = drv.CreateSplashWindow()
	} else {
		w = a.NewWindow("")
	}

	list, entry := createSearchWidget(snippets,
		snippetsText,
		func(snippet *snippet) {
			w.Hide()
			typeSnippet(snippet.content)
		},
		func() {
			w.Hide()
		},
	)

	split := container.NewVSplit(entry, list)
	split.Offset = 0
	w.SetContent(split)
	w.Resize(fyne.NewSize(400, 400))
	w.Canvas().Focus(entry)
	w.CenterOnScreen()

	go listenForHotkeys(w)

	w.ShowAndRun()
}

func loadSnippets() ([]*snippet, []string, error) {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	bytes, err := os.ReadFile(filepath.Join(dir, "snippets.yml"))
	if err != nil {
		return nil, nil, err
	}

	var rawSnippets map[string]string
	err = yaml.Unmarshal(bytes, &rawSnippets)
	if err != nil {
		return nil, nil, err
	}

	var snippets []*snippet
	var snippetsText []string
	for k, v := range rawSnippets {
		snippets = append(snippets, &snippet{
			label:   k,
			content: v,
		})
		snippetsText = append(snippetsText, fmt.Sprintf("%s: %s", k, v))
	}
	return snippets, snippetsText, nil
}

func createSearchWidget(snippets []*snippet, snippetsText []string, onSubmit func(snippet *snippet), onCancel func()) (*widget.List, *snippetEntry) {
	filteredSnippets := snippets

	list := widget.NewList(
		func() int {
			return len(filteredSnippets)
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
			label.SetText(filteredSnippets[id].label)
			content.Text = strings.ReplaceAll(filteredSnippets[id].content, "\n", "\\n")
			ellipsis(container, content)
		},
	)

	var selectedID widget.ListItemID = -1
	list.OnSelected = func(id widget.ListItemID) {
		selectedID = id
	}
	list.Select(0)

	entry := newSnippetEntry()

	resetSearch := func() {
		entry.Text = ""
		entry.OnChanged(entry.Text)
		list.Select(0)
	}

	entry.onTypedKey = func(key *fyne.KeyEvent) {
		if key.Name == "Down" {
			list.Select((selectedID + 1) % len(filteredSnippets))
		} else if key.Name == "Up" {
			list.Select((len(filteredSnippets) + selectedID - 1) % len(filteredSnippets))
		} else if key.Name == "Return" {
			if selectedID >= 0 && selectedID < len(filteredSnippets) {
				onSubmit(filteredSnippets[selectedID])
				resetSearch()
			}
		} else if key.Name == "Escape" {
			onCancel()
			resetSearch()
		}
	}
	entry.OnChanged = func(s string) {
		ranked := fuzzy.RankFindFold(s, snippetsText)
		sort.Sort(ranked)
		filteredSnippets = make([]*snippet, 0)
		for _, r := range ranked {
			filteredSnippets = append(filteredSnippets, snippets[r.OriginalIndex])
		}
		list.Refresh()
		list.Select(0)
	}

	return list, entry
}

func listenForHotkeys(w fyne.Window) {
	robotgo.EventHook(hook.KeyDown, []string{"q", "alt"}, func(e hook.Event) {
		w.Show()
		// robotgo.EventEnd()
	})

	s := robotgo.EventStart()
	<-robotgo.EventProcess(s)
}

func typeSnippet(content string) {
	lines := strings.Split(content, "\n")
	first := true
	for _, l := range lines {
		if !first {
			robotgo.MicroSleep(100)
			robotgo.KeyTap("enter")
		}
		robotgo.TypeStr(l)
		first = false
	}
}

type snippetEntry struct {
	widget.Entry
	onTypedKey func(key *fyne.KeyEvent)
}

func newSnippetEntry() *snippetEntry {
	e := &snippetEntry{}
	e.ExtendBaseWidget(e)
	return e
}

func (e *snippetEntry) TypedKey(key *fyne.KeyEvent) {
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

type myTheme struct{}

func (m myTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DarkTheme().Color(name, theme.VariantDark)
}

func (m myTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DarkTheme().Icon(name)
}

func (m myTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DarkTheme().Font(style)
}

func (m myTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNamePadding {
		return 2
	}
	return theme.DarkTheme().Size(name)
}
