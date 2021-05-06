package util

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// Snippet describes a snippet of text.
type Snippet struct {
	Label   string
	Content string
	Args    []string
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

func unmarshalSnippet(key string, rawSnippet interface{}) (*Snippet, error) {
	snippet := &Snippet{
		Label: key,
	}

	switch rv := rawSnippet.(type) {
	case string:
		snippet.Content = rv
	case map[interface{}]interface{}:
		content, ok := rv["content"]
		if !ok {
			return nil, fmt.Errorf("error loading snippet %s: missing 'content' field", key)
		}
		snippet.Content = content.(string)

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
				snippet.Args = append(snippet.Args, arg)
			}
		}
	default:
		return nil, fmt.Errorf("error loading snippet %s: unknown type %T", key, rawSnippet)
	}

	return snippet, nil
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
