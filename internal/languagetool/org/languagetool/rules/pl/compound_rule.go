package pl

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/compounds.txt
var compoundsFS embed.FS

var (
	compoundOnce sync.Once
	compoundData *rules.CompoundRuleData
)

func loadCompoundData() *rules.CompoundRuleData {
	compoundOnce.Do(func() {
		f, err := compoundsFS.Open("data/compounds.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		d, err := rules.NewCompoundRuleData(f, "/pl/compounds.txt")
		if err != nil {
			panic(err)
		}
		compoundData = d
	})
	return compoundData
}

// CompoundRule ports org.languagetool.rules.pl.CompoundRule.
type CompoundRule struct {
	*rules.AbstractCompoundRule
}

func NewCompoundRule(messages map[string]string) *CompoundRule {
	base := &rules.AbstractCompoundRule{
		ID:                          "PL_COMPOUNDS",
		Description:                 "Sprawdza wyrazy z łącznikiem, np. „łapu capu” zamiast „łapu-capu”",
		WithHyphenMessage:           "Ten wyraz pisze się z łącznikiem.",
		WithoutHyphenMessage:        "Ten wyraz pisze się razem (bez spacji ani łącznika).",
		WithOrWithoutHyphenMessage:  "Ten wyraz pisze się z łącznikiem lub bez niego.",
		ShortDesc:                   "Brak łącznika lub zbędny łącznik",
		SentenceStartsWithUpperCase: true,
		Data:                        loadCompoundData(),
	}
	rules.InitCompoundRuleMeta(base, messages)
	return &CompoundRule{AbstractCompoundRule: base}
}

func (r *CompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}
