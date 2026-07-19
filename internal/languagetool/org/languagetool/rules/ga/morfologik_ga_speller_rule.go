package ga

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	taggingga "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ga"
)

const (
	// MorfologikIrishSpellerRuleID ports MorfologikIrishSpellerRule.getId().
	// Java: "MORFOLOGIK_RULE_GA_IE" (not MORFOLOGIK_RULE_GA).
	MorfologikIrishSpellerRuleID = "MORFOLOGIK_RULE_GA_IE"
	// IrishSpellerDict ports MorfologikIrishSpellerRule.getFileName() → RESOURCE_FILENAME.
	// Java: "/ga/hunspell/ga_IE.dict"
	IrishSpellerDict = "/ga/hunspell/ga_IE.dict"
)

// MorfologikIrishSpellerRule ports rules.ga.MorfologikIrishSpellerRule.
// tokenizingPattern("-"); isMisspelled normalizes maths / halfwidth Latin via tagging/ga.Utils.
type MorfologikIrishSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikIrishSpellerRule() *MorfologikIrishSpellerRule {
	base := morfologik.NewMorfologikSpellerRule(
		MorfologikIrishSpellerRuleID, "ga", IrishSpellerDict, nil)
	// Java MorfologikIrishSpellerRule: super.ignoreWordsWithLength = 1
	if base.SpellingCheckRule != nil {
		base.IgnoreWordsWithLength = 1
	}
	r := &MorfologikIrishSpellerRule{MorfologikSpellerRule: base}
	// Wrap IsMisspelled for maths/halfwidth normalization (Java isMisspelled override).
	inner := r.IsMisspelled
	r.IsMisspelled = func(word string) bool {
		return r.irishIsMisspelled(word, inner)
	}
	return r
}

// irishIsMisspelled ports isMisspelled: simplify mathematical / halfwidth before dict check.
func (r *MorfologikIrishSpellerRule) irishIsMisspelled(word string, inner func(string) bool) bool {
	check := word
	if taggingga.IsAllMathsChars(word) {
		check = taggingga.SimplifyMathematical(word)
	} else if taggingga.IsAllHalfWidthChars(word) {
		check = taggingga.HalfwidthLatinToLatin(word)
	}
	if inner != nil {
		return inner(check)
	}
	return false
}

// Match ports parent Match + hyphen tokenizingPattern (like Breton).
func (r *MorfologikIrishSpellerRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if r == nil || r.MorfologikSpellerRule == nil {
		return nil, nil
	}
	base, err := r.MorfologikSpellerRule.Match(sentence)
	if err != nil || len(base) == 0 {
		return base, err
	}
	out := make([]*rules.RuleMatch, 0, len(base))
	for _, m := range base {
		if m == nil {
			continue
		}
		word := matchSurfaceGA(m, sentence)
		if strings.Contains(word, "-") && r.acceptHyphenParts(word) {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}

func (r *MorfologikIrishSpellerRule) acceptHyphenParts(word string) bool {
	parts := strings.Split(word, "-")
	if len(parts) < 2 {
		return false
	}
	any := false
	for _, p := range parts {
		if p == "" {
			continue
		}
		any = true
		if r.wordIsMisspelled(p) {
			return false
		}
	}
	return any
}

func (r *MorfologikIrishSpellerRule) wordIsMisspelled(word string) bool {
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

func matchSurfaceGA(m *rules.RuleMatch, sent *languagetool.AnalyzedSentence) string {
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
