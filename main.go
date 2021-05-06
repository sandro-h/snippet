package main

import (
	"fmt"
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
	"github.com/sandro-h/snippet/util"
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
	SpaceAfter bool   `yaml:"space_after"`
}

type snippet struct {
	label   string
	content string
	args    []string
}

var cfg *config = &config{
	specialChars:    map[string]specialChar{},
	specialCharList: "",
}

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
	w := newWindow(a)
	argWin := newArgWindow(newWindow(a))

	search := newSearchWidget(snippets,
		func(snippet *snippet) {
			if snippet.args != nil {
				w.Hide()
				argWin.showWithArgs(snippet.args, func(vals map[string]string) {
					typeSnippet(instantiateArgs(snippet.content, vals))
				}, func() {
					w.Show()
				})
			} else {
				w.Hide()
				typeSnippet(snippet.content)
			}
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

func newWindow(a fyne.App) fyne.Window {
	if drv, ok := a.Driver().(desktop.Driver); ok {
		return drv.CreateSplashWindow()
	}
	return a.NewWindow("")
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

	var rawSnippets map[string]interface{}
	err = yaml.Unmarshal(bytes, &rawSnippets)
	if err != nil {
		return nil, err
	}

	var snippets []*snippet
	for k, v := range rawSnippets {
		snippet, err := unmarshalSnippet(k, v)
		if err != nil {
			fmt.Println(err)
		} else {
			snippets = append(snippets, snippet)
		}
	}
	return snippets, nil
}

func unmarshalSnippet(key string, rawSnippet interface{}) (*snippet, error) {
	snippet := &snippet{
		label: key,
	}

	switch rv := rawSnippet.(type) {
	case string:
		snippet.content = rv
	case map[interface{}]interface{}:
		content, ok := rv["content"]
		if !ok {
			return nil, fmt.Errorf("error loading snippet %s: missing 'content' field", key)
		}
		snippet.content = content.(string)

		args, ok := rv["args"]
		if ok {
			rawArgList, ok := args.([]interface{})
			if !ok {
				return nil, fmt.Errorf("error loading snippet %s: 'args' field is not a list of strings", key)
			}
			for i, a := range rawArgList {
				arg, ok := a.(string)
				if !ok {
					return nil, fmt.Errorf("error loading snippet %s: 'args[%d]' field is not string", key, i)
				}
				snippet.args = append(snippet.args, arg)
			}
		}
	default:
		return nil, fmt.Errorf("error loading snippet %s: unknown type %T", key, rawSnippet)
	}

	return snippet, nil
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

func instantiateArgs(content string, vals map[string]string) string {
	res := content
	for k, v := range vals {
		res = strings.ReplaceAll(res, "{"+k+"}", v)
	}
	return res
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
		parts := util.SplitSpecials(str, cfg.specialCharList)
		for _, p := range parts {
			if len(p) == 1 && strings.Contains(cfg.specialCharList, p) {
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

	if key.SpaceAfter {
		robotgo.KeyTap("space")
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
