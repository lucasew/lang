package ru

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Reuses compoundsFS embed from russian_compound_rule.go (same data/compounds.txt).

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

func isCyrillicLetter(r rune) bool {
	return r >= 0x0400 && r <= 0x04FF
}

// RussianDashRule ports org.languagetool.rules.ru.RussianDashRule.
type RussianDashRule struct {
	*rules.AbstractDashRule
}

func NewRussianDashRule(messages map[string]string) *RussianDashRule {
	base := &rules.AbstractDashRule{
		ID:               "RU_DASH_RULE",
		CompoundPatterns: loadDashPatterns(),
		Message:          "Использовано тире вместо дефиса.",
		Description:      "Тире вместо дефиса («из — за» вместо «из-за»).",
		IsLetter:         isCyrillicLetter,
	}
	rules.InitDashRuleMeta(base, messages)
	return &RussianDashRule{AbstractDashRule: base}
}

func (r *RussianDashRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractDashRule.Match(sentence)
}
