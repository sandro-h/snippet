package util

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
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
	Args            []SnippetArg
	Copy            CopyMode
}

// SnippetArg defines an argument to be replaced in the snippet.
type SnippetArg struct {
	Name     string
	Resolver ArgResolver
}

// ArgResolver resolves the argument so it can be replaced in the snippet.
type ArgResolver interface {
	Resolve() string
}

// InputArgResolver marks arguments that require user input.
// It doesn't actually handle them, that is delegated to the UI code.
type InputArgResolver struct{}

// Resolve resolves the input argument. In this case it's a NOOP since the UI code handles
// input args.
func (m *InputArgResolver) Resolve() string {
	return ""
}

// RandomNumberArgResolver resolves the argument to a random integer number.
type RandomNumberArgResolver struct {
	min int
	max int
}

// Resolve returns a random integer number within the resolver range [min,max).
func (m *RandomNumberArgResolver) Resolve() string {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	num := m.min + r.Intn(m.max-m.min)
	return strconv.Itoa(num)
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
			switch arg := a.(type) {
			case string:
				snippet.Args = append(snippet.Args, SnippetArg{Name: arg, Resolver: &InputArgResolver{}})
			case map[interface{}]interface{}:
				parsedArg, err := unmarshalComplexArg(arg)
				if err != nil {
					return fmt.Errorf("error loading snippet %s: 'args[%d]' - %s", key, i, err.Error())
				}
				snippet.Args = append(snippet.Args, *parsedArg)
			default:
				return fmt.Errorf("error loading snippet %s: 'args[%d]' field is not string or map", key, i)
			}
		}
	}
	return nil
}

func unmarshalComplexArg(rawArg map[interface{}]interface{}) (*SnippetArg, error) {
	name, ok := rawArg["name"]
	if !ok {
		return nil, fmt.Errorf("arg is missing 'name' field")
	}

	argType, ok := rawArg["type"]
	if !ok {
		return nil, fmt.Errorf("arg is missing 'type' field")
	}

	var resolver ArgResolver
	var err error
	switch argType {
	case "input":
		resolver = &InputArgResolver{}
	case "random":
		resolver, err = unmarshalRandomNumberResolver(rawArg)
	default:
		return nil, fmt.Errorf("unknown type '%s'", argType)
	}

	if err != nil {
		return nil, err
	}

	return &SnippetArg{Name: name.(string), Resolver: resolver}, nil
}

func unmarshalRandomNumberResolver(rawArg map[interface{}]interface{}) (*RandomNumberArgResolver, error) {
	min := 0
	max := 100

	minVal, ok := rawArg["min"]
	if ok {
		min, ok = minVal.(int)
		if !ok {
			return nil, fmt.Errorf("'min' field is not an integer")
		}
	}

	maxVal, ok := rawArg["max"]
	if ok {
		max, ok = maxVal.(int)
		if !ok {
			return nil, fmt.Errorf("'max' field is not an integer")
		}
	}

	return &RandomNumberArgResolver{min, max}, nil
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
