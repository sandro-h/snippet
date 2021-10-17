package typing

import (
	"runtime"
	"strings"

	"github.com/go-vgo/robotgo"
	"github.com/go-vgo/robotgo/clipboard"
	"github.com/sandro-h/snippet/util"
)

// SpecialChar defines a character that has to be typed in a non-standard way.
type SpecialChar struct {
	Character  string `yaml:"character"`
	SpaceAfter bool   `yaml:"space_after"`
	KeySym     int    `yaml:"key_sym"`
}

// Config specifies the behavior when typing.
type Config struct {
	SpecialChars    map[string]SpecialChar
	SpecialCharList string
}

// TypeSnippet types the snippet content by simulating key presses if copy=false, or simulating a copy/paste if copy=true.
func TypeSnippet(content string, copy util.CopyMode, cfg *Config) {
	switch copy {
	case util.CopyModeNormal:
		copyPasteSnippet(content)
		break
	case util.CopyModeShell:
		copyPasteSnippetToShell(content)
		break
	default:
		typeSnippet(content, cfg)
	}
}

func copyPasteSnippet(content string) {
	robotgo.MicroSleep(50)
	robotgo.PasteStr(content)
}

func copyPasteSnippetToShell(content string) {
	robotgo.MicroSleep(50)
	clipboard.WriteAll(content)

	if runtime.GOOS == "darwin" {
		robotgo.KeyTap("v", "shift", "command")
	}

	robotgo.KeyTap("v", "shift", "control")
}

func typeSnippet(content string, cfg *Config) {
	lines := strings.Split(content, "\n")
	first := true
	for _, l := range lines {
		if !first {
			robotgo.MicroSleep(100)
			robotgo.KeyTap("enter")
		}
		typeStr(l, cfg)
		first = false
	}
}

func typeStr(str string, cfg *Config) {
	// robotgo's linux implementation for typing cannot deal with special keys on non-standard keyboard layouts (e.g. Swiss German),
	// so handle such special keys explicitly.
	if runtime.GOOS == "linux" {
		robotgo.StartMultiToggleKey()
		parts := util.SplitSpecials(str, cfg.SpecialCharList)
		for _, p := range parts {
			if len(p) == 1 && strings.Contains(cfg.SpecialCharList, p) {
				typeSpecialKey(cfg.SpecialChars[p])
			} else {
				robotgo.TypeStr(p)
			}
		}
		robotgo.EndMultiToggleKey()
	} else {
		robotgo.TypeStr(str)
	}
}

func typeSpecialKey(key SpecialChar) {
	robotgo.KeysymType(uint32(key.KeySym))

	if key.SpaceAfter {
		robotgo.KeyTap("space")
	}
}
