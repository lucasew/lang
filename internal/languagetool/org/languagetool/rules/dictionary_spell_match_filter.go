package rules

import "strings"

// DictionarySpellMatchFilter ports org.languagetool.rules.DictionarySpellMatchFilter.
// Drops dictionary-based spelling matches fully covered by an accepted phrase.
//
// Positions are Java String indices (UTF-16 code units), same as RuleMatch
// FromPos/ToPos and AhoCorasickDoubleArrayTrie.parseText on Java strings.
type DictionarySpellMatchFilter struct {
	AcceptedPhrases []string
}

func NewDictionarySpellMatchFilter(phrases []string) *DictionarySpellMatchFilter {
	return &DictionarySpellMatchFilter{AcceptedPhrases: append([]string(nil), phrases...)}
}

// UseDictionaryBasedFilter is implemented by rules that should be filtered.
type UseDictionaryBasedFilter interface {
	UseDictionaryBasedFilterForMatches() bool
}

// Filter removes spelling matches that lie entirely inside an accepted phrase hit.
func (f *DictionarySpellMatchFilter) Filter(ruleMatches []*RuleMatch, plainText string) []*RuleMatch {
	if len(f.AcceptedPhrases) == 0 {
		return ruleMatches
	}
	// find all phrase occurrences (simple multi-pass; Aho-Corasick deferred)
	// spans use UTF-16 indices (Java String / Hit.begin/end).
	type span struct{ begin, end int }
	var hits []span
	for _, phrase := range f.AcceptedPhrases {
		if phrase == "" {
			continue
		}
		startByte := 0
		for {
			i := strings.Index(plainText[startByte:], phrase)
			if i < 0 {
				break
			}
			b := startByte + i
			// byte offset → UTF-16 (Java char) index
			begin := utf16Len(plainText[:b])
			end := begin + utf16Len(phrase)
			hits = append(hits, span{begin, end})
			startByte = b + 1
		}
	}
	if len(hits) == 0 {
		return ruleMatches
	}
	var out []*RuleMatch
	for _, match := range ruleMatches {
		if !usesDictFilter(match) {
			out = append(out, match)
			continue
		}
		drop := false
		for _, h := range hits {
			if match.FromPos >= h.begin && match.ToPos <= h.end {
				drop = true
				break
			}
		}
		if !drop {
			out = append(out, match)
		}
	}
	return out
}

func usesDictFilter(match *RuleMatch) bool {
	if match == nil || match.Rule == nil {
		return false
	}
	if u, ok := match.Rule.(UseDictionaryBasedFilter); ok {
		return u.UseDictionaryBasedFilterForMatches()
	}
	// FakeRule / unknown: treat spelling-like by default for tests when tagged
	return false
}

// DictFilterRule is a test helper implementing UseDictionaryBasedFilter.
type DictFilterRule struct {
	ID string
}

func (r *DictFilterRule) GetID() string { return r.ID }
func (r *DictFilterRule) UseDictionaryBasedFilterForMatches() bool {
	return true
}

// GetPhrases groups consecutive dictionary-based spelling matches into phrase keys
// (ports DictionarySpellMatchFilter.getPhrases).
// match positions and substring are UTF-16 (Java text.getPlainText().substring).
func (f *DictionarySpellMatchFilter) GetPhrases(ruleMatches []*RuleMatch, plainText string) map[string][]*RuleMatch {
	phraseToMatches := map[string][]*RuleMatch{}
	// Java Integer.MIN_VALUE so the first match never counts as adjacent.
	const intMin = -1 << 31
	prevToPos := intMin
	var collectedMatches []*RuleMatch
	var collectedTerms []string
	textU16Len := utf16Len(plainText)
	for _, match := range ruleMatches {
		if match == nil || !usesDictFilter(match) {
			continue
		}
		if match.FromPos < 0 || match.ToPos > textU16Len || match.FromPos > match.ToPos {
			continue
		}
		covered := utf16Substring(plainText, match.FromPos, match.ToPos)
		if match.FromPos == prevToPos+1 {
			key := strings.Join(collectedTerms, " ") + " " + covered
			l := append([]*RuleMatch(nil), collectedMatches...)
			l = append(l, match)
			phraseToMatches[key] = l
		} else {
			collectedTerms = collectedTerms[:0]
			collectedMatches = collectedMatches[:0]
		}
		collectedTerms = append(collectedTerms, covered)
		collectedMatches = append(collectedMatches, match)
		prevToPos = match.ToPos
	}
	return phraseToMatches
}
