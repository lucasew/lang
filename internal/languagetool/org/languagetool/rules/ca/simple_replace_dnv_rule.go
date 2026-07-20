package ca

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace_dnv.txt
var dnvFS embed.FS

var (
	dnvOnce sync.Once
	dnvMap  map[string][]string
)

func loadDNV() map[string][]string {
	dnvOnce.Do(func() {
		f, err := dnvFS.Open("data/replace_dnv.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		dnvMap = m
	})
	return dnvMap
}

// SimpleReplaceDNVRule ports org.languagetool.rules.ca.SimpleReplaceDNVRule
// (AbstractSimpleReplaceLemmasRule). Without lemmas, fail closed (no surface invent).
type SimpleReplaceDNVRule struct {
	*AbstractSimpleReplaceLemmasRule
}

func NewSimpleReplaceDNVRule(messages map[string]string) *SimpleReplaceDNVRule {
	base := &AbstractSimpleReplaceLemmasRule{
		ID:          "CA_SIMPLE_REPLACE_DNV",
		LanguageCode:         "ca",
		SubRuleSpecificIDs:   true,
		Description: "Detecta paraules admeses només per l'AVL i proposa suggeriments de canvi",
		ShortMsg:    "Paraula del DNV",
		WrongLemmas: loadDNV(),
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Paraula del DNV (AVL)."
		},
	}
	_ = messages
	return &SimpleReplaceDNVRule{AbstractSimpleReplaceLemmasRule: base}
}

func (r *SimpleReplaceDNVRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil {
		return nil
	}
	return r.AbstractSimpleReplaceLemmasRule.Match(sentence)
}
