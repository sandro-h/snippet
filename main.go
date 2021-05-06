package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"github.com/fsnotify/fsnotify"
	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
	"github.com/sandro-h/snippet/secrets"
	"github.com/sandro-h/snippet/typing"
	"github.com/sandro-h/snippet/ui"
	"github.com/sandro-h/snippet/util"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/yaml.v2"
)

type config struct {
	typing.Config
}

var cfg *config = &config{
	typing.Config{
		SpecialChars:    map[string]typing.SpecialChar{},
		SpecialCharList: "",
	},
}

var doEncrypt = flag.Bool("encrypt", false, "Encrypt a secret")

func main() {
	flag.Parse()
	if *doEncrypt {
		encryptSecretFlow()
		return
	}

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
	snippets, err := util.LoadSnippets(snippetsFile)
	if err != nil {
		panic(err)
	}

	a := app.New()
	a.Settings().SetTheme(&myTheme{})
	w := newWindow(a)
	argWin := ui.NewArgWindow(newWindow(a))
	pwdWin := newWindow(a)

	search := ui.NewSearchWidget(snippets,
		func(snippet *util.Snippet) {
			w.Hide()
			if snippet.Secret != "" {
				typeSecretSnippet(snippet, w, pwdWin)
			} else if snippet.Args != nil {
				argWin.ShowWithArgs(snippet.Args, func(vals map[string]string) {
					typing.TypeSnippet(util.InstantiateArgs(snippet.Content, vals), &cfg.Config)
				}, func() {
					w.Show()
				})
			} else {
				typing.TypeSnippet(snippet.Content, &cfg.Config)
			}
		},
		func() {
			w.Hide()
		},
	)

	go watchSnippets(snippetsFile, func() {
		snippets, err := util.ReloadSnippets(snippetsFile, search.GetSnippets())
		if err != nil {
			log.Println("error reloading snippets.yml:", err)
			return
		}

		search.SetSnippets(snippets)
	})

	split := container.NewVSplit(search.Entry, search.List)
	split.Offset = 0
	w.SetContent(split)
	w.Resize(fyne.NewSize(400, 250))
	w.Canvas().Focus(search.Entry)
	w.CenterOnScreen()

	go listenForHotkeys(w)

	w.ShowAndRun()
}

func encryptSecretFlow() {
	fmt.Print("Secret>")
	secret, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	fmt.Println()
	fmt.Print("Password>")
	password, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	enc, err := secrets.Encrypt(string(secret), string(password))
	if err != nil {
		panic(err)
	}
	fmt.Println()
	fmt.Println(enc)
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
		SpecialCharList []typing.SpecialChar `yaml:"special_chars"`
	}
	err = yaml.Unmarshal(bytes, &rawCfg)
	if err != nil {
		return nil, err
	}

	cfg := config{
		typing.Config{
			SpecialChars:    map[string]typing.SpecialChar{},
			SpecialCharList: "",
		},
	}
	for _, s := range rawCfg.SpecialCharList {
		cfg.SpecialChars[s.Character] = s
		cfg.SpecialCharList += s.Character
	}

	return &cfg, nil
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

func typeSecretSnippet(snippet *util.Snippet, mainWindow fyne.Window, pwdWindow fyne.Window) {
	if snippet.SecretDecrypted == "" {
		ui.ShowPasswordWindow(pwdWindow, "Password for secret "+snippet.Label,
			func(pwd string) {
				var err error
				snippet.SecretDecrypted, err = secrets.Decrypt(snippet.Secret, pwd)
				if err != nil {
					log.Printf("Could not type secret snippet %s: %s", snippet.Label, err)
					return
				}
				snippet.SecretLastUsed = time.Now()
				typing.TypeSnippet(snippet.SecretDecrypted, &cfg.Config)
			},
			func() {
				mainWindow.Show()
			},
		)
	} else {
		snippet.SecretLastUsed = time.Now()
		typing.TypeSnippet(snippet.SecretDecrypted, &cfg.Config)
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
