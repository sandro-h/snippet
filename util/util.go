package util

import (
	"strings"
	"time"
)

// SplitSpecials splits the string according to the individual special characters
// in the specials parameter. The found special characters are also included in the result.
func SplitSpecials(str string, specials string) []string {
	var parts []string
	// Reminder: s and i are indexes in bytes. Means if we iterate over a unicode point (e.g. 2 bytes wide),
	// i jumps by 2.
	s := 0
	for i, c := range str {
		if s < 0 {
			s = i
		}

		if strings.ContainsRune(specials, c) {
			if i > s {
				parts = append(parts, str[s:i])
			}
			parts = append(parts, string(c))
			// Don't set to i+1 because current char could be more than 2 bytes.
			s = -1
		}
	}

	if s > -1 && s < len(str) {
		parts = append(parts, str[s:])
	}

	return parts
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
