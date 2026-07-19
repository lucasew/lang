package rules

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// NumberInWordFilter ports org.languagetool.rules.AbstractNumberInWordFilter.
// IsMisspelled / GetSuggestions gate candidates like Java; without IsMisspelled → fail-closed.
type NumberInWordFilter struct {
	IsMisspelled   func(word string) bool
	GetSuggestions func(word string) []string
}

func NewNumberInWordFilter() *NumberInWordFilter {
	return &NumberInWordFilter{}
}

var numberInWordDigitRE = regexp.MustCompile(`[0-9]`)

// Suggestions returns candidates that pass the speller gate (Java acceptRuleMatch logic).
func (f *NumberInWordFilter) Suggestions(word string) []string {
	if f == nil || !numberInWordDigitRE.MatchString(word) {
		return nil
	}
	miss := f.IsMisspelled
	if miss == nil {
		// Without speller cannot verify "known" forms (fail-closed: invent none).
		return nil
	}
	var out []string
	repl0 := strings.ReplaceAll(word, "0", "o")
	if repl0 != word && !miss(repl0) {
		out = append(out, repl0)
	}
	without := numberInWordDigitRE.ReplaceAllString(word, "")
	if without != "" && without != word && !miss(without) {
		dup := false
		for _, s := range out {
			if s == without {
				dup = true
				break
			}
		}
		if !dup {
			out = append(out, without)
		}
	}
	if len(out) == 0 && f.GetSuggestions != nil {
		out = append(out, f.GetSuggestions(without)...)
	}
	return out
}

// AcceptRuleMatch ports AbstractNumberInWordFilter.acceptRuleMatch.
// Args: word (surface form containing digits).
func (f *NumberInWordFilter) AcceptRuleMatch(match *RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	reps := f.Suggestions(arguments["word"])
	if len(reps) == 0 {
		return nil
	}
	out := NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
	out.ShortMessage = match.ShortMessage
	out.SetSuggestedReplacements(reps)
	return out
}
