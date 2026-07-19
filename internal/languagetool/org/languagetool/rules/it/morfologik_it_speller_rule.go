package it

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const (
	// MorfologikItalianSpellerRuleID ports MorfologikItalianSpellerRule.getId().
	// Java returns "MORFOLOGIK_RULE_IT_IT" (not MORFOLOGIK_RULE_IT).
	MorfologikItalianSpellerRuleID = "MORFOLOGIK_RULE_IT_IT"
	// ItalianSpellerDict ports MorfologikItalianSpellerRule.getFileName() → RESOURCE_FILENAME.
	// Java: "/it/hunspell/it_IT.dict"
	ItalianSpellerDict = "/it/hunspell/it_IT.dict"
)

// MorfologikItalianSpellerRule ports rules.it.MorfologikItalianSpellerRule.
type MorfologikItalianSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikItalianSpellerRule() *MorfologikItalianSpellerRule {
	return &MorfologikItalianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikItalianSpellerRuleID, "it", ItalianSpellerDict, nil),
	}
}

// Match ports parent Match + orderSuggestions (capitalized-dup drop).
func (r *MorfologikItalianSpellerRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if r == nil || r.MorfologikSpellerRule == nil {
		return nil, nil
	}
	base, err := r.MorfologikSpellerRule.Match(sentence)
	if err != nil || len(base) == 0 {
		return base, err
	}
	for _, m := range base {
		if m == nil {
			continue
		}
		word := matchSurfaceIT(m, sentence)
		sugs := m.GetSuggestedReplacements()
		if len(sugs) == 0 {
			continue
		}
		m.SetSuggestedReplacements(orderItalianSuggestions(sugs, word))
	}
	return base, nil
}

// orderItalianSuggestions ports MorfologikItalianSpellerRule.orderSuggestions:
// drop capitalized suggestion when original word is not capitalized and the
// lowercase form is also present in the suggestion list.
func orderItalianSuggestions(suggestions []string, word string) []string {
	if len(suggestions) == 0 {
		return suggestions
	}
	// originalSuggestionsStr for membership checks
	set := make(map[string]struct{}, len(suggestions))
	for _, s := range suggestions {
		set[s] = struct{}{}
	}
	out := make([]string, 0, len(suggestions))
	for _, sug := range suggestions {
		// !isCapitalizedWord(word) && isCapitalizedWord(sug) && list contains sug.toLowerCase()
		if !tools.IsCapitalizedWord(word) && tools.IsCapitalizedWord(sug) {
			if _, ok := set[strings.ToLower(sug)]; ok {
				continue
			}
		}
		out = append(out, sug)
	}
	return out
}

func matchSurfaceIT(m *rules.RuleMatch, sent *languagetool.AnalyzedSentence) string {
	if m == nil || sent == nil {
		return ""
	}
	text := sent.GetText()
	from, to := m.GetFromPos(), m.GetToPos()
	if from < 0 || from >= to {
		return ""
	}
	runes := []rune(text)
	if to <= len(runes) {
		return string(runes[from:to])
	}
	if to <= len(text) {
		return text[from:to]
	}
	return ""
}
