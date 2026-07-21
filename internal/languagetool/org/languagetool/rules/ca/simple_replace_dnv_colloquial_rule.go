package ca

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace_dnv_colloquial.txt
var dnvColloquialFS embed.FS

var (
	dnvColloquialOnce sync.Once
	dnvColloquialMap  map[string][]string
)

func loadDNVColloquial() map[string][]string {
	dnvColloquialOnce.Do(func() {
		f, err := dnvColloquialFS.Open("data/replace_dnv_colloquial.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		dnvColloquialMap = m
	})
	return dnvColloquialMap
}

// SimpleReplaceDNVColloquialRule ports org.languagetool.rules.ca.SimpleReplaceDNVColloquialRule
// (AbstractSimpleReplaceLemmasRule). Without lemmas, fail closed (no surface invent).
type SimpleReplaceDNVColloquialRule struct {
	*AbstractSimpleReplaceLemmasRule
}

func NewSimpleReplaceDNVColloquialRule(messages map[string]string) *SimpleReplaceDNVColloquialRule {
	base := &AbstractSimpleReplaceLemmasRule{
		ID:                 "CA_SIMPLE_REPLACE_DNV_COLLOQUIAL",
		LanguageCode:       "ca",
		SubRuleSpecificIDs: true,
		Description:        "Detecta paraules marcades com a col·loquials en el DNV.",
		ShortMsg:           "Paraula o expressió col·loquial.",
		WrongLemmas:        loadDNVColloquial(),
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Paraula o expressió col·loquial."
		},
	}
	_ = messages
	return &SimpleReplaceDNVColloquialRule{AbstractSimpleReplaceLemmasRule: base}
}

func (r *SimpleReplaceDNVColloquialRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil {
		return nil
	}
	return r.AbstractSimpleReplaceLemmasRule.Match(sentence)
}
