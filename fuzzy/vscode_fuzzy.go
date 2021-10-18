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
	"sort"
	"strings"
)

const noMatch = 0
const maxTargetLen = 512

type fuzzyScore struct {
	Score          int
	MatchPositions []int
}

type fuzzyScoreWithRanges struct {
	Score       int
	MatchRanges []MatchRange
}

type byScore []Match

func (a byScore) Len() int           { return len(a) }
func (a byScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byScore) Less(i, j int) bool { return a[i].Score >= a[j].Score }

type byRangeStart []MatchRange

func (a byRangeStart) Len() int           { return len(a) }
func (a byRangeStart) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byRangeStart) Less(i, j int) bool { return a[i].Start <= a[j].Start }

// FindFuzzy finds targets that fuzzily match the query.
func FindFuzzy(query string, targets []string) []Match {
	var matches []Match
	query = strings.TrimSpace(query)
	queryLower := strings.ToLower(query)
	for i, t := range targets {
		score := scoreFuzzySingleOrMultiple(t, query, queryLower, true)
		if score.Score > 0 {
			m := Match{Str: t, Index: i, Score: score.Score, MatchedRanges: score.MatchRanges}
			matches = append(matches, m)
		}
	}
	sort.Stable(byScore(matches))
	return matches
}

func scoreFuzzySingleOrMultiple(target string, query string, queryLower string, allowNonContiguousMatches bool) fuzzyScoreWithRanges {
	queryParts := strings.Split(query, " ")
	if len(queryParts) > 1 {
		queryPartsLower := strings.Split(queryLower, " ")
		return scoreFuzzyMultiple(target, queryParts, queryPartsLower, allowNonContiguousMatches)
	}

	return scoreFuzzySingle(target, query, queryLower, allowNonContiguousMatches)
}

func scoreFuzzyMultiple(target string, queryParts []string, queryPartsLower []string, allowNonContiguousMatches bool) fuzzyScoreWithRanges {
	noScore := fuzzyScoreWithRanges{Score: 0}
	var scores []int
	var matchRanges []MatchRange
	for i, q := range queryParts {
		score := scoreFuzzySingle(target, q, queryPartsLower[i], allowNonContiguousMatches)
		if score.Score == noMatch {
			// if a single query value does not match, return with
			// no score entirely, we require all queries to match
			return noScore
		}
		scores = append(scores, score.Score)
		matchRanges = append(matchRanges, score.MatchRanges...)
	}

	mergedScore := fuzzyScoreWithRanges{Score: 0}
	for _, s := range scores {
		mergedScore.Score += s
	}
	mergedScore.MatchRanges = normalizeMatchRanges(matchRanges)

	return mergedScore
}

func scoreFuzzySingle(target string, query string, queryLower string, allowNonContiguousMatches bool) fuzzyScoreWithRanges {
	score := scoreFuzzy(target, query, queryLower, allowNonContiguousMatches)
	return fuzzyScoreWithRanges{Score: score.Score, MatchRanges: mergeMatchPositions(score.MatchPositions)}
}

func scoreFuzzy(target string, query string, queryLower string, allowNonContiguousMatches bool) fuzzyScore {
	noScore := fuzzyScore{0, []int{}}

	if target == "" || query == "" {
		return noScore // return early if target or query are undefined
	}

	targetLength := minInt(maxTargetLen, len(target))
	queryLength := len(query)

	if targetLength < queryLength {
		return noScore // impossible for query to be contained in target
	}

	targetLower := strings.ToLower(target)
	res := doScoreFuzzy(query, queryLower, queryLength, target, targetLower, targetLength, allowNonContiguousMatches)

	return res
}

func doScoreFuzzy(query string, queryLower string, queryLength int, target string, targetLower string, targetLength int, allowNonContiguousMatches bool) fuzzyScore {
	scores := make([]int, queryLength*targetLength)
	matches := make([]int, queryLength*targetLength)

	//
	// Build Scorer Matrix:
	//
	// The matrix is composed of query q and target t. For each index we score
	// q[i] with t[i] and compare that with the previous score. If the score is
	// equal or larger, we keep the match. In addition to the score, we also keep
	// the length of the consecutive matches to use as boost for the score.
	//
	//      t   a   r   g   e   t
	//  q
	//  u
	//  e
	//  r
	//  y
	//
	for queryIndex := 0; queryIndex < queryLength; queryIndex++ {
		queryIndexOffset := queryIndex * targetLength
		queryIndexPreviousOffset := queryIndexOffset - targetLength

		queryIndexGtNull := queryIndex > 0

		queryCharAtIndex := query[queryIndex]
		queryLowerCharAtIndex := queryLower[queryIndex]

		for targetIndex := 0; targetIndex < targetLength; targetIndex++ {
			targetIndexGtNull := targetIndex > 0

			currentIndex := queryIndexOffset + targetIndex
			leftIndex := currentIndex - 1
			diagIndex := queryIndexPreviousOffset + targetIndex - 1

			leftScore := 0
			if targetIndexGtNull {
				leftScore = scores[leftIndex]
			}
			diagScore := 0
			if queryIndexGtNull && targetIndexGtNull {
				diagScore = scores[diagIndex]
			}

			matchesSequenceLength := 0
			if queryIndexGtNull && targetIndexGtNull {
				matchesSequenceLength = matches[diagIndex]
			}

			// If we are not matching on the first query character any more, we only produce a
			// score if we had a score previously for the last query index (by looking at the diagScore).
			// This makes sure that the query always matches in sequence on the target. For example
			// given a target of "ede" and a query of "de", we would otherwise produce a wrong high score
			// for query[1] ("e") matching on target[0] ("e") because of the "beginning of word" boost.
			var score int
			if diagScore == 0 && queryIndexGtNull {
				score = 0
			} else {
				score = computeCharScore(queryCharAtIndex, queryLowerCharAtIndex, target, targetLower, targetIndex, matchesSequenceLength)
			}

			// We have a score and its equal or larger than the left score
			// Match: sequence continues growing from previous diag value
			// Score: increases by diag score value
			isValidScore := score > 0 && diagScore+score >= leftScore
			if isValidScore && (
			// We don't need to check if it's contiguous if we allow non-contiguous matches
			allowNonContiguousMatches ||
				// We must be looking for a contiguous match.
				// Looking at an index higher than 0 in the query means we must have already
				// found out this is contiguous otherwise there wouldn't have been a score
				queryIndexGtNull ||
				// lastly check if the query is completely contiguous at this index in the target
				strings.HasPrefix(targetLower[targetIndex:], queryLower)) {
				matches[currentIndex] = matchesSequenceLength + 1
				scores[currentIndex] = diagScore + score
			} else {
				// We either have no score or the score is lower than the left score
				// Match: reset to 0
				// Score: pick up from left hand side
				matches[currentIndex] = noMatch
				scores[currentIndex] = leftScore
			}
		}
	}

	// Restore Positions (starting from bottom right of matrix)
	var positions []int
	queryIndex := queryLength - 1
	targetIndex := targetLength - 1
	for queryIndex >= 0 && targetIndex >= 0 {
		currentIndex := queryIndex*targetLength + targetIndex
		match := matches[currentIndex]
		if match == noMatch {
			targetIndex-- // go left
		} else {
			positions = append(positions, targetIndex)

			// go up and left
			queryIndex--
			targetIndex--
		}
	}

	reverse(positions)
	return fuzzyScore{scores[queryLength*targetLength-1], positions}
}

func computeCharScore(queryCharAtIndex byte, queryLowerCharAtIndex byte, target string, targetLower string, targetIndex int, matchesSequenceLength int) int {
	score := 0

	if !considerAsEqual(queryLowerCharAtIndex, targetLower[targetIndex]) {
		return score // no match of characters
	}

	// Character match bonus
	score++

	// Consecutive match bonus
	if matchesSequenceLength > 0 {
		score += (matchesSequenceLength * 5)
	}

	// Same case bonus
	if queryCharAtIndex == target[targetIndex] {
		score++
	}

	// Start of word bonus
	if targetIndex == 0 {
		score += 8
	} else {

		// After separator bonus
		separatorBonus := scoreSeparatorAtPos(target[targetIndex-1])
		if separatorBonus > 0 {
			score += separatorBonus
		} else if isUpper(target[targetIndex]) {
			// Inside word upper case bonus (camel case)
			score += 2
		}
	}

	return score
}

func considerAsEqual(a byte, b byte) bool {
	if a == b {
		return true
	}

	// Special case path separators: ignore platform differences
	if a == '/' || a == '\\' {
		return b == '/' || b == '\\'
	}

	return false
}

func scoreSeparatorAtPos(charCode byte) int {
	switch charCode {
	case '/':
		fallthrough
	case '\\':
		return 5 // prefer path separators...
	case '_':
		fallthrough
	case '-':
		fallthrough
	case '.':
		fallthrough
	case ' ':
		fallthrough
	case '\'':
		fallthrough
	case '"':
		fallthrough
	case ':':
		return 4 // ...over other separators
	}
	return 0
}

func isUpper(code byte) bool {
	return 'A' <= code && code <= 'Z'
}

func reverse(s []int) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func normalizeMatchRanges(matches []MatchRange) []MatchRange {

	// sort matches by start to be able to normalize
	sort.Stable(byRangeStart(matches))

	// merge matches that overlap
	var normalizedMatches []MatchRange
	var currentMatch *MatchRange
	for _, match := range matches {

		// if we have no current match or the matches
		// do not overlap, we take it as is and remember
		// it for future merging
		if currentMatch == nil || !matchOverlaps(currentMatch, &match) {
			normalizedMatches = append(normalizedMatches, match)
			currentMatch = &normalizedMatches[len(normalizedMatches)-1]
		} else {
			// otherwise we merge the matches
			currentMatch.Start = minInt(currentMatch.Start, match.Start)
			currentMatch.End = maxInt(currentMatch.End, match.End)
		}
	}

	return normalizedMatches
}

func matchOverlaps(matchA *MatchRange, matchB *MatchRange) bool {
	if matchA.End < matchB.Start {
		return false // A ends before B starts
	}

	if matchB.End < matchA.Start {
		return false // B ends before A starts
	}

	return true
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a int, b int) int {
	if a >= b {
		return a
	}
	return b
}
