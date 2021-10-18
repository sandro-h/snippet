package fuzzy

import "sort"

// MatchRange describes the start and end index of a match in the target string.
type MatchRange struct {
	Start int
	End   int
}

// Match describes a fuzzy search match.
type Match struct {
	// The matched string.
	Str string
	// The index of the matched string in the supplied slice.
	Index int
	// The indexes of matched ranges. Useful for highlighting matches.
	MatchedRanges []MatchRange
	// Score used to rank matches
	Score int
}

// MultiMatch describes a match in one or both target lists by SearchFuzzyMulti.
type MultiMatch struct {
	Index  int
	Score  int
	Match1 Match
	Match2 Match
}

type byIndex []Match
type byMultiScore []MultiMatch

func (a byIndex) Len() int           { return len(a) }
func (a byIndex) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byIndex) Less(i, j int) bool { return a[i].Index < a[j].Index }

func (a byMultiScore) Len() int           { return len(a) }
func (a byMultiScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byMultiScore) Less(i, j int) bool { return a[i].Score >= a[j].Score }

// SearchFuzzyMulti searches for source in multiple target lists using a fuzzy
// algorithm. Matches with the same index in both targets are merged. If not, a resulting
// match may only contain a match in the first or second target list.
// This is meant to be used to search a list of objects where multiple fields of an object
// should be searched.
// The result is ordered from best to worst fitting match.
func SearchFuzzyMulti(source string, targets1 []string, targets2 []string) []MultiMatch {
	NullMatch := Match{Index: -1}

	matches1 := SearchFuzzy(source, targets1)
	matches2 := SearchFuzzy(source, targets2)
	sort.Stable(byIndex(matches1))
	sort.Stable(byIndex(matches2))
	var combined []MultiMatch

	k2 := 0
	for _, m1 := range matches1 {
		for k2 < len(matches2) && matches2[k2].Index < m1.Index {
			addMultiMatch(&combined, NullMatch, matches2[k2])
			k2++
		}

		if k2 < len(matches2) && matches2[k2].Index == m1.Index {
			addMultiMatch(&combined, m1, matches2[k2])
			k2++
		} else {
			addMultiMatch(&combined, m1, NullMatch)
		}
	}

	for ; k2 < len(matches2); k2++ {
		addMultiMatch(&combined, NullMatch, matches2[k2])
	}

	sort.Stable(byMultiScore(combined))
	return combined
}

func addMultiMatch(combined *[]MultiMatch, m1 Match, m2 Match) {
	index := 0
	score := 0
	if m1.Index > -1 {
		index = m1.Index
		score += m1.Score
	}
	if m2.Index > -1 {
		index = m2.Index
		score += m2.Score
	}
	*combined = append(*combined, MultiMatch{
		Index:  index,
		Match1: m1,
		Match2: m2,
		Score:  score,
	})
}

// SearchFuzzy searches for source in the list of targets using a fuzzy
// algorithm. The result is ordered from best to worst fitting match.
func SearchFuzzy(source string, targets []string) []Match {
	if source == "" {
		var res []Match
		for i, t := range targets {
			res = append(res, Match{
				Str:   t,
				Index: i,
				Score: 0,
			})
		}
		return res
	}

	return FindFuzzy(source, targets)
}

func mergeMatchPositions(positions []int) []MatchRange {
	var ranges []MatchRange
	var cur *MatchRange
	for _, i := range positions {
		if cur == nil {
			cur = &MatchRange{Start: i, End: i}
		} else {
			if i == cur.End+1 {
				cur.End = i
			} else {
				ranges = append(ranges, *cur)
				cur = &MatchRange{Start: i, End: i}
			}
		}
	}
	if cur != nil {
		ranges = append(ranges, *cur)
	}
	return ranges
}
