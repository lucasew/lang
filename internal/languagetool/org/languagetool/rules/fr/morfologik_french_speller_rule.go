package fr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

const (
	// MorfologikFrenchSpellerRuleID ports MorfologikFrenchSpellerRule.getId().
	// Java returns "FR_SPELLING_RULE" (not the generic MORFOLOGIK_RULE_FR form).
	MorfologikFrenchSpellerRuleID = "FR_SPELLING_RULE"
	// FrenchSpellerDict ports MorfologikFrenchSpellerRule.getFileName() → DICT_FILE.
	// Java: "/fr/french.dict" (not /fr/hunspell/fr.dict).
	FrenchSpellerDict = "/fr/french.dict"
)

// MorfologikFrenchSpellerRule ports rules.fr.MorfologikFrenchSpellerRule.
// orderSuggestions + additionalTop (units, camelCase, digit split, apostrophe/hyphen
// via TagPOS). Full elision rewrite still partial without FrenchTagger POS fidelity.
type MorfologikFrenchSpellerRule struct {
	*morfologik.MorfologikSpellerRule
	// TagPOS ports FrenchTagger.tag for findSuggestion POS gates (fail-closed when nil).
	TagPOS func(word string) []string
}

func NewMorfologikFrenchSpellerRule() *MorfologikFrenchSpellerRule {
	r := &MorfologikFrenchSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikFrenchSpellerRuleID, "fr", FrenchSpellerDict, nil),
	}
	// Java MorfologikFrenchSpellerRule ctor: setIgnoreTaggedWords().
	r.IgnoreTaggedWords = true
	return r
}

// Match ports parent Match + orderSuggestions + additional top suggestions.
func (r *MorfologikFrenchSpellerRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
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
		word := matchSurfaceFR(m, sentence)
		var top []string
		if t := additionalTopFrenchSuggestions(word); len(t) > 0 {
			top = append(top, t...)
		} else if parts := splitCamelCase(word); len(parts) > 1 && tokenizers.UTF16Len(parts[0]) > 1 {
			ok := true
			for _, p := range parts {
				if r.wordIsMisspelled(p) {
					ok = false
					break
				}
			}
			if ok {
				top = append(top, strings.Join(parts, " "))
			}
		}
		if len(top) == 0 {
			if ds := r.digitSplitTopSuggestion(word); ds != "" {
				top = append(top, ds)
			}
		}
		if len(top) == 0 {
			top = r.apostropheHyphenTopSuggestions(word)
		}
		sugs := m.GetSuggestedReplacements()
		if len(top) > 0 {
			sugs = append(top, sugs...)
		}
		if len(sugs) > 0 {
			m.SetSuggestedReplacements(orderFrenchSuggestions(sugs, word))
		}
	}
	return base, nil
}

func (r *MorfologikFrenchSpellerRule) wordIsMisspelled(word string) bool {
	if r == nil || word == "" {
		return false
	}
	if r.IsMisspelled != nil {
		return r.IsMisspelled(word)
	}
	if r.Speller != nil {
		return r.Speller.IsMisspelled(word)
	}
	return false
}

func matchSurfaceFR(m *rules.RuleMatch, sent *languagetool.AnalyzedSentence) string {
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

// UseInOffice ports useInOffice() — force-enable in LO/OO extension.
func (r *MorfologikFrenchSpellerRule) UseInOffice() bool { return true }
