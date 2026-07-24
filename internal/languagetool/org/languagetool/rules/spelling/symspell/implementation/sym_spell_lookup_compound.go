package implementation

import (
	"math"
	"sort"
	"strings"
)

// LookupCompound ports SymSpell.lookupCompound(input) — uses maxDictionaryEditDistance.
func (s *SymSpell) LookupCompound(input string) []SuggestItem {
	if s == nil {
		return nil
	}
	return s.LookupCompoundMax(input, s.maxDictionaryEditDistance)
}

// LookupCompoundMax ports SymSpell.lookupCompound(input, maxEditDistance).
func (s *SymSpell) LookupCompoundMax(input string, maxEditDistance int) []SuggestItem {
	return s.lookupCompoundMax(input, maxEditDistance)
}

func (s *SymSpell) lookupCompoundMax(input string, maxEditDistance int) []SuggestItem {
	if s == nil {
		return nil
	}
	if maxEditDistance > s.maxDictionaryEditDistance {
		panic("Dist to big " + itoa(maxEditDistance))
	}
	termList1 := parseWords(input)

	// translate every term to its best suggestion, otherwise it remains unchanged
	lastCombi := false
	suggestionParts := make([]SuggestItem, 0, len(termList1))

	for i := 0; i < len(termList1); i++ {
		suggestions := s.LookupMax(termList1[i], VerbosityTop, maxEditDistance)

		// combi check, always before split. i > 0 because we can't split on zero.
		if i > 0 && !lastCombi {
			suggestionsCombi := s.LookupMax(termList1[i-1]+termList1[i], VerbosityTop, maxEditDistance)
			if len(suggestionsCombi) > 0 {
				best1 := suggestionParts[len(suggestionParts)-1]
				var best2 SuggestItem
				if len(suggestions) > 0 {
					best2 = suggestions[0]
				} else {
					// No suggestion -> it might be correct?
					best2 = NewSuggestItem(termList1[i], maxEditDistance+1, 0)
				}
				editDistance := NewEditDistance(termList1[i-1]+" "+termList1[i], Damerau)
				if suggestionsCombi[0].Distance+1 < editDistance.DamerauLevenshteinDistance(best1.Term+" "+best2.Term, maxEditDistance) {
					// suggestionsCombi.get(0).distance++; suggestionParts.set(last, combi)
					combi := suggestionsCombi[0]
					combi.Distance++
					suggestionParts[len(suggestionParts)-1] = combi
					lastCombi = true
					continue
				}
			}
		}

		lastCombi = false

		// always split terms without suggestion / never split terms with suggestion ed=0 / never split single char terms
		if len(suggestions) > 0 && (suggestions[0].Distance == 0 || javaStringLen(termList1[i]) == 1) {
			suggestionParts = append(suggestionParts, suggestions[0])
		} else {
			// if no perfect suggestion, split word into pairs
			var suggestionsSplit []SuggestItem
			// add original term
			if len(suggestions) > 0 {
				suggestionsSplit = append(suggestionsSplit, suggestions[0])
			}

			termLen := javaStringLen(termList1[i])
			if termLen > 1 {
				for j := 1; j < termLen; j++ {
					part1 := javaSubstring(termList1[i], 0, j)
					part2 := javaSubstring(termList1[i], j, termLen)
					suggestions1 := s.LookupMax(part1, VerbosityTop, maxEditDistance)
					if len(suggestions1) == 0 {
						continue
					}
					// suggestion top = split_1 suggestion top
					if len(suggestions) > 0 && suggestions[0].EqualTerm(suggestions1[0]) {
						continue
					}
					suggestions2 := s.LookupMax(part2, VerbosityTop, maxEditDistance)
					if len(suggestions2) == 0 {
						continue
					}
					if len(suggestions) > 0 && suggestions[0].EqualTerm(suggestions2[0]) {
						continue
					}
					// select best suggestion for split pair
					split := suggestions1[0].Term + " " + suggestions2[0].Term
					editDistance := NewEditDistance(termList1[i], Damerau)
					dist := editDistance.DamerauLevenshteinDistance(split, maxEditDistance)
					count := suggestions1[0].Count
					if suggestions2[0].Count < count {
						count = suggestions2[0].Count
					}
					suggestionSplit := NewSuggestItem(split, dist, count)
					if suggestionSplit.Distance >= 0 {
						suggestionsSplit = append(suggestionsSplit, suggestionSplit)
					}
					// early termination of split
					if suggestionSplit.Distance == 1 {
						break
					}
				}

				if len(suggestionsSplit) > 0 {
					sort.SliceStable(suggestionsSplit, func(a, b int) bool {
						return suggestionsSplit[a].Less(suggestionsSplit[b])
					})
					suggestionParts = append(suggestionParts, suggestionsSplit[0])
				} else {
					// Java: new SuggestItem(term, 0, maxEditDistance + 1)
					suggestionParts = append(suggestionParts, NewSuggestItem(termList1[i], 0, int64(maxEditDistance+1)))
				}
			} else {
				suggestionParts = append(suggestionParts, NewSuggestItem(termList1[i], 0, int64(maxEditDistance+1)))
			}
		}
	}

	// Join parts
	suggestion := NewSuggestItem("", math.MaxInt32, math.MaxInt64)
	var b strings.Builder
	for _, si := range suggestionParts {
		b.WriteString(si.Term)
		b.WriteByte(' ')
		if si.Count < suggestion.Count {
			suggestion.Count = si.Count
		}
	}
	// replaceAll("\\s+$", "")
	suggestion.Term = strings.TrimRight(b.String(), " \t\n\r")
	editDistance := NewEditDistance(suggestion.Term, Damerau)
	// Java: DamerauLevenshteinDistance(input, maxDictionaryEditDistance)
	suggestion.Distance = editDistance.DamerauLevenshteinDistance(input, s.maxDictionaryEditDistance)
	return []SuggestItem{suggestion}
}
