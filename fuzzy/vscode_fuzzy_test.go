package fuzzy

/*
	Most of this code is taken and translated to Golang from the vscode fuzzyScorer.
	All credit and admiration to the original authors.

	The original license notice from vscode:

	MIT License

	Copyright (c) 2015 - present Microsoft Corporation

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE.
*/

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScoreFuzzy(t *testing.T) {
	target := "HeLlo-World"

	scores := [...]fuzzyScore{
		_doScore(target, "HelLo-World", true), // direct case match
		_doScore(target, "hello-world", true), // direct mix-case match
		_doScore(target, "HW", true),          // direct case prefix (multiple)
		_doScore(target, "hw", true),          // direct mix-case prefix (multiple)
		_doScore(target, "H", true),           // direct case prefix
		_doScore(target, "h", true),           // direct mix-case prefix
		_doScore(target, "W", true),           // direct case word prefix
		_doScore(target, "Ld", true),          // in-string case match (multiple)
		_doScore(target, "ld", true),          // in-string mix-case match (consecutive, avoids scattered hit)
		_doScore(target, "w", true),           // direct mix-case word prefix
		_doScore(target, "L", true),           // in-string case match
		_doScore(target, "l", true),           // in-string mix-case match
		_doScore(target, "4", true),           // no match
	}

	lastScore := 100000
	for _, s := range scores {
		assert.LessOrEqual(t, s.Score, lastScore)
		lastScore = s.Score
	}
}

func _doScore(target string, query string, allowNonContiguousMatches bool) fuzzyScore {
	return scoreFuzzy(target, query, strings.ToLower(query), allowNonContiguousMatches)
}
