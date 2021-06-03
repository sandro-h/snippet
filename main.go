package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

type hotkeyConfig struct {
	activateHotkeys []string
	editorHotkeys   []string
	editorCmd       string
}

type config struct {
	typing.Config
	secretTTL time.Duration
	hotkeyConfig
}

const defaultSecretTTL = 10 * time.Minute

var defaultActivateHotkeys = []string{"q", "alt"}

var defaultEditorHotkeys = []string{"e", "alt"}

var cfg *config = &config{
	Config: typing.Config{
		SpecialChars:    map[string]typing.SpecialChar{},
		SpecialCharList: "",
	},
	secretTTL: defaultSecretTTL,
	hotkeyConfig: hotkeyConfig{
		activateHotkeys: defaultActivateHotkeys,
		editorHotkeys:   defaultEditorHotkeys,
	},
}

type appState struct {
	snippets []*util.Snippet
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

	state := &appState{}

	snippetsFile := filepath.Join(dir, "snippets.yml")
	if _, err := os.Stat(snippetsFile); os.IsNotExist(err) {
		os.Create(snippetsFile)
	}
	var err error
	state.snippets, err = util.LoadSnippets(snippetsFile)
	if err != nil {
		panic(err)
	}

	a := app.New()
	a.Settings().SetTheme(&myTheme{})
	w := newWindow(a)
	argWin := ui.NewArgWindow(newWindow(a))
	pwdWin := newWindow(a)

	search := ui.NewSearchWidget(state.snippets,
		func(snippet *util.Snippet) {
			w.Hide()
			if snippet.Secret != "" {
				typeSecretSnippet(snippet, w, pwdWin)
			} else if snippet.Args != nil {
				typeArgSnippet(snippet, w, argWin)

			} else {
				typing.TypeSnippet(snippet.Content, snippet.Copy, &cfg.Config)
			}
		},
		func() {
			w.Hide()
		},
	)

	go watchSnippets(snippetsFile, func() {
		snippets, err := util.ReloadSnippets(snippetsFile, state.snippets)
		if err != nil {
			log.Println("error reloading snippets.yml:", err)
			return
		}

		// Heuristic for fsnotify double-loads where the first load can't read anything.
		// That would destroy the old runtime state if we don't catch it here.
		if len(snippets) == 0 {
			return
		}

		state.snippets = snippets
		search.SetSnippets(snippets)
	})

	split := container.NewVSplit(search.Entry, search.List)
	split.Offset = 0
	w.SetContent(split)
	w.Resize(fyne.NewSize(400, 250))
	w.Canvas().Focus(search.Entry)
	w.CenterOnScreen()

	go listenForHotkeys(w, snippetsFile, cfg.hotkeyConfig)
	go periodicallyEvictSecrets(state, cfg.secretTTL)

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
		SecretTTL       string               `yaml:"secret_ttl"`
		EditorCmd       string               `yaml:"editor_cmd"`
		ActivateHotkeys []string             `yaml:"activate_hotkeys"`
		EditorHotkeys   []string             `yaml:"editor_hotkeys"`
	}
	err = yaml.Unmarshal(bytes, &rawCfg)
	if err != nil {
		return nil, err
	}

	cfg := config{
		Config: typing.Config{
			SpecialChars:    map[string]typing.SpecialChar{},
			SpecialCharList: "",
		},
		secretTTL: defaultSecretTTL,
		hotkeyConfig: hotkeyConfig{
			editorCmd: rawCfg.EditorCmd,
		},
	}

	if rawCfg.ActivateHotkeys != nil {
		cfg.activateHotkeys = rawCfg.ActivateHotkeys
	} else {
		cfg.activateHotkeys = defaultActivateHotkeys
	}

	if rawCfg.EditorHotkeys != nil {
		cfg.editorHotkeys = rawCfg.EditorHotkeys
	} else {
		cfg.editorHotkeys = defaultEditorHotkeys
	}

	if rawCfg.SecretTTL != "" {
		dur, err := time.ParseDuration(rawCfg.SecretTTL)
		if err != nil {
			return nil, err
		}
		cfg.secretTTL = dur
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

func listenForHotkeys(w fyne.Window, snippetsFile string, hotkeyCfg hotkeyConfig) {
	robotgo.EventHook(hook.KeyDown, hotkeyCfg.activateHotkeys, func(e hook.Event) {
		w.Show()
	})

	if hotkeyCfg.editorCmd != "" {
		editorCmdParts := strings.Split(hotkeyCfg.editorCmd, " ")
		editorCmdParts = append(editorCmdParts, snippetsFile)
		robotgo.EventHook(hook.KeyDown, hotkeyCfg.editorHotkeys, func(e hook.Event) {
			cmd := exec.Command(editorCmdParts[0], editorCmdParts[1:]...)
			err := cmd.Start()
			if err != nil {
				log.Printf("Could not run %s: %s", cmd, err)
			}
		})
	}

	s := robotgo.EventStart()
	<-robotgo.EventProcess(s)
}

func typeArgSnippet(snippet *util.Snippet, mainWindow fyne.Window, argWin *ui.ArgWindow) {
	argWin.ShowWithArgs(snippet.Args, func(vals map[string]string) {
		typing.TypeSnippet(util.InstantiateArgs(snippet.Content, vals), snippet.Copy, &cfg.Config)
	}, func() {
		mainWindow.Show()
	})
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
				typing.TypeSnippet(snippet.SecretDecrypted, snippet.Copy, &cfg.Config)
			},
			func() {
				mainWindow.Show()
			},
		)
	} else {
		snippet.SecretLastUsed = time.Now()
		typing.TypeSnippet(snippet.SecretDecrypted, snippet.Copy, &cfg.Config)
	}
}

func periodicallyEvictSecrets(state *appState, ttl time.Duration) {
	for {
		now := time.Now()
		for _, s := range state.snippets {
			if s.SecretDecrypted != "" && now.Sub(s.SecretLastUsed) > ttl {
				s.SecretDecrypted = ""
			}
		}

		// Use 30s interval by default, except if ttl/2 is lower than that.
		// But do max 1 check per second.
		time.Sleep(util.MaxDur(1*time.Second, util.MinDur(30*time.Second, ttl/2)))
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
