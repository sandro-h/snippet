package util

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// CopyMode describes whether a snippet is copy-pasted instead of typed and in what fashion.
type CopyMode int

const (
	// CopyModeNone uses regular typing instead of copy-pasting.
	CopyModeNone CopyMode = iota
	// CopyModeNormal uses the standard Ctrl+V shortcut to copy-paste the snippet
	CopyModeNormal = iota
	// CopyModeShell uses the Ctrl+Shift+V shortcut to copy-paste the snippet into a terminal, where
	// Ctrl+V usually doesn't work.
	CopyModeShell = iota
)

// Snippet describes a snippet of text.
type Snippet struct {
	Label           string
	Content         string
	Secret          string
	SecretDecrypted string
	SecretLastUsed  time.Time
	Args            []string
	Copy            CopyMode
}

// LoadSnippets loads a list of snippets from a YAML file.
func LoadSnippets(snippetsFile string) ([]*Snippet, error) {
	bytes, err := os.ReadFile(snippetsFile)
	if err != nil {
		return nil, err
	}

	var rawSnippets map[string]interface{}
	err = yaml.Unmarshal(bytes, &rawSnippets)
	if err != nil {
		return nil, err
	}

	var snippets []*Snippet
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

// ReloadSnippets reloads the snippets (usually when snippetsFile content changed),
// and transfers any runtime data of the old snippets to the matching new snippets.
func ReloadSnippets(snippetsFile string, oldSnippets []*Snippet) ([]*Snippet, error) {
	newSnippets, err := LoadSnippets(snippetsFile)
	if err != nil {
		return nil, err
	}

	// Transfer runtime snippet data to newly loaded snippets.
	oldSnippetsMap := make(map[string]*Snippet)
	for _, s := range oldSnippets {
		oldSnippetsMap[s.Label] = s
	}

	for _, s := range newSnippets {
		if os, ok := oldSnippetsMap[s.Label]; ok {
			s.SecretDecrypted = os.SecretDecrypted
			s.SecretLastUsed = os.SecretLastUsed
		}
	}
	return newSnippets, nil
}

func unmarshalSnippet(key string, rawSnippet interface{}) (*Snippet, error) {
	snippet := &Snippet{
		Label: key,
	}

	switch rv := rawSnippet.(type) {
	case string:
		snippet.Content = rv
	case map[interface{}]interface{}:
		err := unmarshalContent(key, rv, snippet)
		if err != nil {
			return nil, err
		}
		err = unmarshalArguments(key, rv, snippet)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("error loading snippet %s: unknown type %T", key, rawSnippet)
	}

	return snippet, nil
}

func unmarshalContent(key string, rawValue map[interface{}]interface{}, snippet *Snippet) error {
	var ok bool
	content, hasContent := rawValue["content"]
	secret, hasSecret := rawValue["secret"]
	if hasContent {
		snippet.Content, ok = content.(string)
		if !ok {
			return fmt.Errorf("error loading snippet %s: 'content' field is not string", key)
		}
	} else if hasSecret {
		snippet.Content = "******"
		snippet.Secret, ok = secret.(string)
		if !ok {
			return fmt.Errorf("error loading snippet %s: 'secret' field is not string", key)
		}
	} else {
		return fmt.Errorf("error loading snippet %s: missing 'content' or 'secret' field", key)
	}

	copy, hasCopy := rawValue["copy"]
	if hasCopy {
		copyStr, ok := copy.(string)
		if ok {
			switch copyStr {
			case "none":
				snippet.Copy = CopyModeNone
				break
			case "normal":
				snippet.Copy = CopyModeNormal
				break
			case "shell":
				snippet.Copy = CopyModeShell
				break
			default:
				snippet.Copy = CopyModeNone
				ok = false
			}
		}

		if !ok {
			fmt.Printf("Warning: snippet %s - 'copy' field should be one of: none, normal, shell. Ignoring field.\n", key)
		}
	}

	return nil
}

func unmarshalArguments(key string, rawValue map[interface{}]interface{}, snippet *Snippet) error {
	args, ok := rawValue["args"]
	if ok {
		rawArgList, ok := args.([]interface{})
		if !ok {
			return fmt.Errorf("error loading snippet %s: 'args' field is not a list of strings", key)
		}
		for i, a := range rawArgList {
			arg, ok := a.(string)
			if !ok {
				return fmt.Errorf("error loading snippet %s: 'args[%d]' field is not string", key, i)
			}
			snippet.Args = append(snippet.Args, arg)
		}
	}
	return nil
}

// InstantiateArgs takes a snippet content and a map of argument names to values and replaces
// all instances of {arg} with the corresponding value in the map.
func InstantiateArgs(content string, vals map[string]string) string {
	res := content
	for k, v := range vals {
		res = strings.ReplaceAll(res, "{"+k+"}", v)
	}
	return res
}
