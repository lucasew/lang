package br

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/compounds.txt
var compoundsFS embed.FS

var (
	compoundsOnce sync.Once
	compoundsData *rules.CompoundRuleData
)

func loadCompounds() *rules.CompoundRuleData {
	compoundsOnce.Do(func() {
		f, err := compoundsFS.Open("data/compounds.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		d, err := rules.NewCompoundRuleData(f, "/br/compounds.txt")
		if err != nil {
			panic(err)
		}
		compoundsData = d
	})
	return compoundsData
}

// BretonCompoundRule ports org.languagetool.rules.br.BretonCompoundRule.
type BretonCompoundRule struct {
	*rules.AbstractCompoundRule
}

func NewBretonCompoundRule(messages map[string]string) *BretonCompoundRule {
	base := &rules.AbstractCompoundRule{
		Messages:                   messages,
		ID:                         "BR_COMPOUNDS",
		Description:                "Mots composés",
		WithHyphenMessage:          "Skrivet e vez ar ger-mañ boaz gant ur varrennig-stagañ.",
		WithoutHyphenMessage:       "Ar ger-mañ a zo skrivet boaz evel unan hepken.",
		WithOrWithoutHyphenMessage: "An droienn-mañ a zo skrivet evel ur ger hepken pe gant ur varrennig-stagañ.",
		ShortDesc:                  "Kudenn barrennig-stagañ",
		Data:                       loadCompounds(),
	}
	return &BretonCompoundRule{AbstractCompoundRule: base}
}

func (r *BretonCompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}
