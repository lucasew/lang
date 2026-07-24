package ar

import (
	"regexp"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/hunspell"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const (
	ArabicHunspellRuleID   = "HUNSPELL_RULE_AR"
	ArabicHunspellDictPath = "/ar/hunspell/ar.dic"
)

// arTokenSplit splits on non-letters excluding tashkeel (ArabicHunspellSpellerRule.tokenizeText).
var arTokenSplit = regexp.MustCompile(`[^\p{L}` + tools.TashkeelChars + `]+`)

// ArabicHunspellSpellerRule ports org.languagetool.rules.ar.ArabicHunspellSpellerRule.
type ArabicHunspellSpellerRule struct {
	*hunspell.HunspellRule
}

func NewArabicHunspellSpellerRule(dict hunspell.HunspellDictionary) *ArabicHunspellSpellerRule {
	base := hunspell.NewHunspellRule("ar", dict)
	base.ID = ArabicHunspellRuleID
	base.Description = "Possible spelling mistake"
	// Java isLatinScript() = false
	if base.SpellingCheckRule != nil {
		base.NonLatinScript = true
	}
	r := &ArabicHunspellSpellerRule{HunspellRule: base}
	base.IsMisspelled = r.IsMisspelledStripped
	return r
}

func (r *ArabicHunspellSpellerRule) GetID() string { return ArabicHunspellRuleID }

func (r *ArabicHunspellSpellerRule) GetDictFilenameInResources(_ string) string {
	return ArabicHunspellDictPath
}

func (r *ArabicHunspellSpellerRule) IsLatinScript() bool { return false }

// TokenizeArabicSpellText ports tokenizeText.
func TokenizeArabicSpellText(sentence string) []string {
	parts := arTokenSplit.Split(sentence, -1)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// IsMisspelledStripped strips tashkeel before dictionary lookup.
func (r *ArabicHunspellSpellerRule) IsMisspelledStripped(word string) bool {
	if r == nil || r.HunspellRule == nil {
		return false
	}
	return r.HunspellRule.IsMisspelledWord(tools.RemoveTashkeel(word))
}

func (r *ArabicHunspellSpellerRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if r == nil {
		return nil, nil
	}
	r.HunspellRule.IsMisspelled = r.IsMisspelledStripped
	return r.HunspellRule.Match(sentence)
}

func HasArabicLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}
