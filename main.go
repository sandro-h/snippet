package main

import (
	"image/color"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"github.com/fsnotify/fsnotify"
	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
	"gopkg.in/yaml.v2"
)

type config struct {
	specialChars    map[string]specialChar
	specialCharList string
}

type specialChar struct {
	Character  string `yaml:"character"`
	Key        string `yaml:"key"`
	CommandKey string `yaml:"command"`
}

type snippet struct {
	label   string
	content string
}

var cfg *config

func main() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	configFile := filepath.Join(dir, "config.yml")
	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		cfg, err = loadConfig(configFile)
		if err != nil {
			panic(err)
		}
	}

	snippetsFile := filepath.Join(dir, "snippets.yml")
	if _, err := os.Stat(snippetsFile); os.IsNotExist(err) {
		os.Create(snippetsFile)
	}
	snippets, err := loadSnippets(snippetsFile)
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

	search := newSearchWidget(snippets,
		func(snippet *snippet) {
			w.Hide()
			typeSnippet(snippet.content)
		},
		func() {
			w.Hide()
		},
	)

	go watchSnippets(snippetsFile, func() {
		snippets, err := loadSnippets(snippetsFile)
		if err != nil {
			log.Println("error reloading snippets.yml:", err)
			return
		}
		search.setSnippets(snippets)
	})

	split := container.NewVSplit(search.entry, search.list)
	split.Offset = 0
	w.SetContent(split)
	w.Resize(fyne.NewSize(400, 250))
	w.Canvas().Focus(search.entry)
	w.CenterOnScreen()

	go listenForHotkeys(w)

	w.ShowAndRun()
}

func loadConfig(configFile string) (*config, error) {
	bytes, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var rawCfg struct {
		SpecialCharList []specialChar `yaml:"special_chars"`
	}
	err = yaml.Unmarshal(bytes, &rawCfg)
	if err != nil {
		return nil, err
	}

	cfg := config{
		specialChars:    map[string]specialChar{},
		specialCharList: "",
	}
	for _, s := range rawCfg.SpecialCharList {
		cfg.specialChars[s.Character] = s
		cfg.specialCharList += s.Character
	}

	return &cfg, nil
}

func loadSnippets(snippetsFile string) ([]*snippet, error) {
	bytes, err := os.ReadFile(snippetsFile)
	if err != nil {
		return nil, err
	}

	var rawSnippets map[string]string
	err = yaml.Unmarshal(bytes, &rawSnippets)
	if err != nil {
		return nil, err
	}

	var snippets []*snippet
	for k, v := range rawSnippets {
		snippets = append(snippets, &snippet{
			label:   k,
			content: v,
		})
	}
	return snippets, nil
}

func watchSnippets(snippetsFile string, onModified func()) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	err = watcher.Add(snippetsFile)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				onModified()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
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
		typeStr(l)
		first = false
	}
}

func typeStr(str string) {
	// robotgo's linux implementation for typing cannot deal with special keys on non-standard keyboard layouts (e.g. Swiss German),
	// so handle such special keys explicitly.
	if runtime.GOOS == "linux" {
		parts := splitSpecials(str, cfg.specialCharList)
		for _, p := range parts {
			if len(p) == 1 && strings.Index(cfg.specialCharList, p) > -1 {
				typeSpecialKey(cfg.specialChars[p])
			} else {
				robotgo.TypeStr(p)
			}
		}
	} else {
		robotgo.TypeStr(str)
	}
}

func typeSpecialKey(key specialChar) {
	if key.CommandKey == "gralt" {
		robotgo.KeyToggle("gralt", "down")
		robotgo.KeyTap(key.Key)
		robotgo.KeyToggle("gralt", "up")
	} else {
		robotgo.KeyTap(key.Key, key.CommandKey)
	}
}

func splitSpecials(str string, specials string) []string {
	var parts []string
	s := 0
	e := 0
	for i, c := range str {
		if strings.IndexRune(specials, c) > -1 {
			if e > s {
				parts = append(parts, str[s:e])
			}
			parts = append(parts, string(c))
			s = i + 1
			e = s
		} else {
			e++
		}
	}

	if e > s {
		parts = append(parts, str[s:e])
	}

	return parts
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
