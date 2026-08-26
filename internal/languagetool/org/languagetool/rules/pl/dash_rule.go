package pl

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

var (
	dashOnce     sync.Once
	dashPatterns []string
)

func loadDashPatterns() []string {
	dashOnce.Do(func() {
		f, err := compoundsFS.Open("data/compounds.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		p, err := rules.LoadDashCompoundPatterns(f)
		if err != nil {
			panic(err)
		}
		dashPatterns = p
	})
	return dashPatterns
}

// DashRule ports org.languagetool.rules.pl.DashRule.
// Compounds written with en/em dashes instead of hyphens.
type DashRule struct {
	*rules.AbstractDashRule
}

func NewDashRule(messages map[string]string) *DashRule {
	base := &rules.AbstractDashRule{
		ID:               "DASH_RULE",
		CompoundPatterns: loadDashPatterns(),
		Message:          "Błędne użycie myślnika zamiast łącznika.",
		Description:      "Sprawdza, czy wyrazy pisane z łącznikiem zapisano z myślnikami (np. „Lądek — Zdrój” zamiast „Lądek-Zdrój”).",
	}
	rules.InitDashRuleMeta(base, messages)
	return &DashRule{AbstractDashRule: base}
}

func (r *DashRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractDashRule.Match(sentence)
}
