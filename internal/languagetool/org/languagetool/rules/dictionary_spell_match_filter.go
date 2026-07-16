package rules

import "strings"

// DictionarySpellMatchFilter ports org.languagetool.rules.DictionarySpellMatchFilter.
// Drops dictionary-based spelling matches fully covered by an accepted phrase.
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
	type span struct{ begin, end int }
	var hits []span
	for _, phrase := range f.AcceptedPhrases {
		if phrase == "" {
			continue
		}
		start := 0
		for {
			i := strings.Index(plainText[start:], phrase)
			if i < 0 {
				break
			}
			b := start + i
			hits = append(hits, span{b, b + len(phrase)})
			start = b + 1
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
func (f *DictionarySpellMatchFilter) GetPhrases(ruleMatches []*RuleMatch, plainText string) map[string][]*RuleMatch {
	phraseToMatches := map[string][]*RuleMatch{}
	// Java Integer.MIN_VALUE so the first match never counts as adjacent.
	const intMin = -1 << 31
	prevToPos := intMin
	var collectedMatches []*RuleMatch
	var collectedTerms []string
	for _, match := range ruleMatches {
		if match == nil || !usesDictFilter(match) {
			continue
		}
		if match.FromPos < 0 || match.ToPos > len(plainText) || match.FromPos > match.ToPos {
			continue
		}
		covered := plainText[match.FromPos:match.ToPos]
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
