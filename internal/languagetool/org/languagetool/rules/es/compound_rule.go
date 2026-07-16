package es

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
		d, err := rules.NewCompoundRuleData(f, "/es/compounds.txt")
		if err != nil {
			panic(err)
		}
		compoundData = d
	})
	return compoundData
}

// CompoundRule ports org.languagetool.rules.es.CompoundRule.
type CompoundRule struct {
	*rules.AbstractCompoundRule
}

func NewCompoundRule(messages map[string]string) *CompoundRule {
	base := &rules.AbstractCompoundRule{
		Messages:                    messages,
		ID:                          "ES_COMPOUNDS",
		Description:                 "Palabras compuestas con guion: $match",
		WithHyphenMessage:           "Se escribe con un guion.",
		WithoutHyphenMessage:        "Se escribe junto sin espacio ni guion.",
		WithOrWithoutHyphenMessage:  "Se escribe junto o con un guion.",
		ShortDesc:                   "Error de palabra compuesta",
		SentenceStartsWithUpperCase: true,
		Data:                        loadCompoundData(),
	}
	base.UseSubRuleSpecificIDs()
	return &CompoundRule{AbstractCompoundRule: base}
}

func (r *CompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}
