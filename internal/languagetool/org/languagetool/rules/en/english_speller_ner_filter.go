package en

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/languagemodel"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ner"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CountProvider is the GetCount surface used by AbstractEnglishSpellerRule.NER filter
// (Java BaseLanguageModel.getCount).
type CountProvider interface {
	GetCountToken(token string) int64
	GetCount(tokens []string) int64
}

// filterNERMatches ports AbstractEnglishSpellerRule.filter (NER + LM named-entity skip).
// Matches covered by a PERSON span may be dropped when suggestions look rare vs covered.
func filterNERMatches(matches []*rules.RuleMatch, sentenceText string, namedEntities []ner.Span, lm CountProvider) []*rules.RuleMatch {
	if len(matches) == 0 || lm == nil || len(namedEntities) == 0 {
		return matches
	}
	toFilter := map[*rules.RuleMatch]bool{}
	for _, neSpan := range namedEntities {
		for _, match := range matches {
			if match == nil {
				continue
			}
			// Java: neSpan.getStart() <= match.getFromPos() && neSpan.getEnd() >= match.getToPos()
			if neSpan.GetStart() > match.GetFromPos() || neSpan.GetEnd() < match.GetToPos() {
				continue
			}
			covered := safeByteSlice(sentenceText, match.GetFromPos(), match.GetToPos())
			if !tools.StartsWithUppercase(covered) {
				continue
			}
			textCount := lm.GetCountToken(covered)
			var mostCommonRepl string
			mostCommonReplCount := textCount
			i := 0
			nonZeroReplacements := 0
			lookupFailures := 0
			translations := 0
			objs := match.GetSuggestedReplacementObjects()
			if len(objs) == 0 {
				for _, s := range match.GetSuggestedReplacements() {
					objs = append(objs, rules.NewSuggestedReplacement(s))
				}
			}
			for _, repl := range objs {
				if repl == nil {
					continue
				}
				if repl.GetType() == rules.SuggestionTypeTranslation {
					translations++
				}
				replList := strings.Fields(repl.GetReplacement())
				if len(replList) == 0 {
					// empty — skip count
				} else if len(replList) <= 3 {
					// hard-coding 3grams is not good, but a base LM doesn't know about ngrams...
					replCount := lm.GetCount(replList)
					if replCount > 0 {
						nonZeroReplacements++
					}
					if replCount > mostCommonReplCount {
						mostCommonRepl = repl.GetReplacement()
						mostCommonReplCount = replCount
					}
				} else {
					lookupFailures++
				}
				if i++; i >= 4 {
					break
				}
			}
			_ = mostCommonReplCount
			if translations == 0 && nonZeroReplacements == 0 && lookupFailures == 0 {
				// e.g. "Fastow", which only offers zero-count multi-token suggestions
				toFilter[match] = true
			} else if translations == 0 && mostCommonRepl != "" {
				dist := enLevenshtein(mostCommonRepl, covered)
				if dist > 2 {
					toFilter[match] = true
				}
			}
		}
	}
	if len(toFilter) == 0 {
		return matches
	}
	out := make([]*rules.RuleMatch, 0, len(matches))
	for _, m := range matches {
		if !toFilter[m] {
			out = append(out, m)
		}
	}
	return out
}

// AsCountProvider adapts *languagemodel.BaseLanguageModel or any CountProvider.
func AsCountProvider(v any) CountProvider {
	if v == nil {
		return nil
	}
	if c, ok := v.(CountProvider); ok {
		return c
	}
	if m, ok := v.(*languagemodel.BaseLanguageModel); ok && m != nil && m.Counts != nil {
		return m.Counts
	}
	return nil
}

func safeByteSlice(s string, from, to int) string {
	if from < 0 {
		from = 0
	}
	if to > len(s) {
		to = len(s)
	}
	if from >= to {
		return ""
	}
	return s[from:to]
}

// enLevenshtein ports Apache Commons LevenshteinDistance.apply (rune-level).
func enLevenshtein(a, b string) int {
	if a == b {
		return 0
	}
	ra := []rune(a)
	rb := []rune(b)
	la, lb := len(ra), len(rb)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	cur := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		cur[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			ins := cur[j-1] + 1
			del := prev[j] + 1
			sub := prev[j-1] + cost
			m := ins
			if del < m {
				m = del
			}
			if sub < m {
				m = sub
			}
			cur[j] = m
		}
		prev, cur = cur, prev
	}
	return prev[lb]
}
