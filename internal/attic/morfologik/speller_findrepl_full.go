package morfologik

import (
	"sort"
	"strings"
)

const (
	upperSearchLimitFindRepl = 15
	minWordLengthFindRepl    = 4
	maxReplRecursion         = 6
)

// FindReplacementCandidates ports Speller.findReplacementCandidates(word, false).
// Uses Dictionary .info: input/output conversion, theRest, short anyToOne/Two, diacritics.
func (d *Dictionary) FindReplacementCandidates(word string, maxEdit int) []CandidateData {
	return d.findReplacementCandidates(word, maxEdit, false)
}

// FindSimilarWordCandidates ports Speller.findSimilarWordCandidates (evenIfInDictionary).
func (d *Dictionary) FindSimilarWordCandidates(word string, maxEdit int) []CandidateData {
	return d.findReplacementCandidates(word, maxEdit, true)
}

func (d *Dictionary) findReplacementCandidates(word string, maxEdit int, evenIfWordInDictionary bool) []CandidateData {
	if d == nil || d.FSA == nil || word == "" {
		return nil
	}
	if maxEdit < 1 {
		maxEdit = 1
	}
	word = applyConversionPairs(word, d.InputConversion)
	if len(word) == 0 || len(word) >= MaxWordLength {
		return nil
	}
	if d.Contains(word) && !evenIfWordInDictionary {
		return nil
	}

	fsaSp := NewSpellerFSA(d, maxEdit)
	fsaSp.IgnoreDiacritics = d.IgnoreDiacritics
	fsaSp.ConvertCase = d.ConvertCase
	fsaSp.EquivalentChars = d.EquivalentChars
	if len(d.ReplacementShort) > 0 {
		pairs := make([]struct{ From, To string }, len(d.ReplacementShort))
		for i, p := range d.ReplacementShort {
			pairs[i].From, pairs[i].To = p.From, p.To
		}
		fsaSp.LoadReplacementPairs(pairs)
	}

	var wordsToCheck []string
	var raw []CandidateData
	if d.ReplacementTheRest != nil && d.ReplacementTheRest.Len() > 0 && len(word) > 1 {
		for _, wordChecked := range getAllReplacements(word, d.ReplacementTheRest, 0, 0) {
			if d.Contains(wordChecked) {
				raw = append(raw, fsaSp.MakeCandidateData(wordChecked, 0))
			} else {
				// Java: toLowerCase/toUpperCase(dictionaryMetadata.getLocale())
				low := d.ToLower(wordChecked)
				up := d.ToUpper(wordChecked)
				if d.Contains(low) {
					raw = append(raw, fsaSp.MakeCandidateData(low, 0))
				}
				if d.Contains(up) {
					raw = append(raw, fsaSp.MakeCandidateData(up, 0))
				}
				if len(low) > 1 {
					// Java: Character.toUpperCase(lowerWord.charAt(0)) + lowerWord.substring(1)
					firstUp := d.initialUppercase(low)
					if d.Contains(firstUp) {
						raw = append(raw, fsaSp.MakeCandidateData(firstUp, 0))
					}
				}
			}
			wordsToCheck = append(wordsToCheck, wordChecked)
		}
	} else {
		wordsToCheck = []string{word}
	}

	fsaSp.ResetHMatrix()
	i := 1
	for _, wordChecked := range wordsToCheck {
		i++
		if i > upperSearchLimitFindRepl {
			break
		}
		if len([]rune(wordChecked)) < minWordLengthFindRepl && i > 2 {
			break
		}
		fsaSp.AppendFindRepl(&raw, wordChecked)
	}

	sort.SliceStable(raw, func(a, b int) bool {
		return raw[a].Distance < raw[b].Distance
	})
	seen := map[string]struct{}{}
	out := make([]CandidateData, 0, len(raw))
	for _, cd := range raw {
		replaced := applyConversionPairs(cd.Word, d.OutputConversion)
		if replaced == "" || replaced == word {
			continue
		}
		if _, ok := seen[replaced]; ok {
			continue
		}
		seen[replaced] = struct{}{}
		// Java: new CandidateData(replaced, cd.origDistance)
		out = append(out, fsaSp.MakeCandidateData(replaced, cd.OrigDistance))
	}
	return out
}

// getAllReplacements ports Speller.getAllReplacements (theRest only).
func getAllReplacements(str string, theRest *OrderedStringListMap, fromIndex, level int) []string {
	if theRest == nil || theRest.Len() == 0 {
		return []string{str}
	}
	if level > maxReplRecursion {
		return []string{str}
	}
	sb := str
	index := MaxWordLength
	key := ""
	keyLength := 0
	found := false
	strippedKeyForSelected := ""
	for _, auxKey := range theRest.Keys {
		startAnchor := strings.HasPrefix(auxKey, "^")
		endAnchor := strings.HasSuffix(auxKey, "$")
		stripped := auxKey
		if startAnchor || endAnchor {
			stripped = stripAnchorsMeta(auxKey)
		}
		auxIndex := -1
		if startAnchor && fromIndex > 0 {
			continue
		} else if startAnchor {
			if strings.HasPrefix(sb, stripped) {
				auxIndex = 0
			}
		} else if endAnchor {
			expectedIndex := len(sb) - len(stripped)
			if expectedIndex >= fromIndex && expectedIndex >= 0 && strings.HasSuffix(sb, stripped) {
				auxIndex = expectedIndex
			}
		} else {
			if i := strings.Index(sb[fromIndex:], auxKey); i >= 0 {
				auxIndex = fromIndex + i
			}
		}
		if auxIndex > -1 && (auxIndex < index || (auxIndex == index && !(len(stripped) < keyLength))) {
			index = auxIndex
			key = auxKey
			keyLength = len(stripped)
			strippedKeyForSelected = stripped
		}
	}
	var replaced []string
	if index < MaxWordLength {
		for _, rep := range theRest.Get(key) {
			if !found {
				replaced = append(replaced, getAllReplacements(str, theRest, index+len(strippedKeyForSelected), level+1)...)
				found = true
			}
			ind := -1
			searchFrom := fromIndex - len(rep) + 1
			if searchFrom < 0 {
				searchFrom = 0
			}
			if i := strings.Index(sb[searchFrom:], rep); i >= 0 {
				ind = searchFrom + i
			}
			if len(rep) > len(strippedKeyForSelected) && ind > -1 &&
				(ind == index || ind == index-len(rep)+1) {
				continue
			}
			newStr := sb[:index] + rep + sb[index+len(strippedKeyForSelected):]
			replaced = append(replaced, getAllReplacements(newStr, theRest, index+len(rep), level+1)...)
		}
	}
	if !found {
		replaced = append(replaced, sb)
	}
	return replaced
}

func stripAnchorsMeta(key string) string {
	start := 0
	end := len(key)
	if strings.HasPrefix(key, "^") {
		start = 1
	}
	if strings.HasSuffix(key, "$") {
		end--
	}
	if start >= end {
		return ""
	}
	return key[start:end]
}
