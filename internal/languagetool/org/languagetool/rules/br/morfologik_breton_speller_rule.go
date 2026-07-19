package br

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

const (
	MorfologikBretonSpellerRuleID   = "MORFOLOGIK_RULE_BR_FR"
	MorfologikBretonSpellerRuleDict = "/br/hunspell/br_FR.dict"
)

// MorfologikBretonSpellerRule ports rules.br.MorfologikBretonSpellerRule.
// tokenizingPattern = "-" (hyphen-split spellcheck); setIgnoreTaggedWords.
type MorfologikBretonSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikBretonSpellerRule() *MorfologikBretonSpellerRule {
	r := &MorfologikBretonSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikBretonSpellerRuleID, "br", MorfologikBretonSpellerRuleDict, nil),
	}
	// Java MorfologikBretonSpellerRule ctor: setIgnoreTaggedWords().
	r.IgnoreTaggedWords = true
	return r
}

// Match ports parent Match with tokenizingPattern("-"):
// if the surface contains '-', only flag when a hyphen-separated part is misspelled
// (Java splits and getRuleMatches each segment).
func (r *MorfologikBretonSpellerRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
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
		word := matchSurfaceBR(m, sentence)
		if strings.Contains(word, "-") {
			// Java tokenizingPattern: spell each side of '-'; if all parts accepted, drop match.
			if r.acceptHyphenParts(word) {
				continue
			}
		}
		out = append(out, m)
	}
	return out, nil
}

func (r *MorfologikBretonSpellerRule) acceptHyphenParts(word string) bool {
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

func (r *MorfologikBretonSpellerRule) wordIsMisspelled(word string) bool {
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

func matchSurfaceBR(m *rules.RuleMatch, sent *languagetool.AnalyzedSentence) string {
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
