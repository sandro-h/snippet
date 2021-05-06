package util

import (
	"strings"
	"time"

	"github.com/sahilm/fuzzy"
)

// SplitSpecials splits the string according to the individual special characters
// in the specials parameter. The found special characters are also included in the result.
func SplitSpecials(str string, specials string) []string {
	var parts []string
	s := 0
	e := 0
	for i, c := range str {
		if strings.ContainsRune(specials, c) {
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

// SearchFuzzy searches for source in the list of targets using a fuzzy
// algorithm. The result is ordered from best to worst fitting match.
func SearchFuzzy(source string, targets []string) fuzzy.Matches {
	if source == "" {
		var res fuzzy.Matches
		for i, t := range targets {
			res = append(res, fuzzy.Match{
				Str:   t,
				Index: i,
				Score: 0,
			})
		}
		return res
	}
	return fuzzy.Find(source, targets)
}

// MinDur returns the smaller of two durations.
func MinDur(d1 time.Duration, d2 time.Duration) time.Duration {
	if d1 < d2 {
		return d1
	}
	return d2
}

// MaxDur returns the larger of two durations.
func MaxDur(d1 time.Duration, d2 time.Duration) time.Duration {
	if d1 > d2 {
		return d1
	}
	return d2
}
