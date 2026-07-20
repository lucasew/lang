package ca

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace_dnv_secondary.txt
var dnvSecondaryFS embed.FS

var (
	dnvSecondaryOnce sync.Once
	dnvSecondaryMap  map[string][]string
)

func loadDNVSecondary() map[string][]string {
	dnvSecondaryOnce.Do(func() {
		f, err := dnvSecondaryFS.Open("data/replace_dnv_secondary.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		// Lemma keys as in file (Java loadFromPath); case-sensitive lemmas from tagger.
		dnvSecondaryMap = m
	})
	return dnvSecondaryMap
}

// SimpleReplaceDNVSecondaryRule ports org.languagetool.rules.ca.SimpleReplaceDNVSecondaryRule
// (AbstractSimpleReplaceLemmasRule: lemma + optional Catalan synthesizer).
// Without lemma POS readings, fail closed (no surface invent of dispost/plurals).
type SimpleReplaceDNVSecondaryRule struct {
	*AbstractSimpleReplaceLemmasRule
}

func NewSimpleReplaceDNVSecondaryRule(messages map[string]string) *SimpleReplaceDNVSecondaryRule {
	base := &AbstractSimpleReplaceLemmasRule{
		ID:          "CA_SIMPLE_REPLACE_DNV_SECONDARY",
		LanguageCode:         "ca",
		SubRuleSpecificIDs:   true,
		Description: "Recomana paraules o formes preferents.",
		ShortMsg:    "Forma secundària",
		WrongLemmas: loadDNVSecondary(),
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Paraula o forma secundària."
		},
	}
	_ = messages
	return &SimpleReplaceDNVSecondaryRule{AbstractSimpleReplaceLemmasRule: base}
}

// Match delegates to lemma path.
func (r *SimpleReplaceDNVSecondaryRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil {
		return nil
	}
	return r.AbstractSimpleReplaceLemmasRule.Match(sentence)
}
