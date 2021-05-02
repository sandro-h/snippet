package util

import "strings"

// SplitSpecials splits the string according to the individual special characters
// in the specials parameter. The found special characters are also included in the result.
func SplitSpecials(str string, specials string) []string {
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
