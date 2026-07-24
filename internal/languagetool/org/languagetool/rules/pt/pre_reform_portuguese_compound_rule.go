package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/pre-reform-compounds.txt
var preReformFS embed.FS

var (
	preReformOnce sync.Once
	preReformData *rules.CompoundRuleData
)

func loadPreReformCompounds() *rules.CompoundRuleData {
	preReformOnce.Do(func() {
		f, err := preReformFS.Open("data/pre-reform-compounds.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		d, err := rules.NewCompoundRuleData(f, "/pt/pre-reform-compounds.txt")
		if err != nil {
			panic(err)
		}
		preReformData = d
	})
	return preReformData
}

// PreReformPortugueseCompoundRule ports org.languagetool.rules.pt.PreReformPortugueseCompoundRule.
type PreReformPortugueseCompoundRule struct {
	*rules.AbstractCompoundRule
}

func NewPreReformPortugueseCompoundRule(messages map[string]string) *PreReformPortugueseCompoundRule {
	base := &rules.AbstractCompoundRule{
		Messages:                   messages,
		ID:                         "PT_COMPOUNDS_PRE_REFORM",
		Description:                "Palavras compostas",
		WithHyphenMessage:          "Esta palavra é hifenizada.",
		WithoutHyphenMessage:       "Esta palavra é composta por justaposição.",
		WithOrWithoutHyphenMessage: "Esta palavra pode ser composta por justaposição ou hifenizada.",
		ShortDesc:                  "Este conjunto forma uma palavra composta.",
		Data:                       loadPreReformCompounds(),
	}
	base.UseSubRuleSpecificIDs()
	return &PreReformPortugueseCompoundRule{AbstractCompoundRule: base}
}

func (r *PreReformPortugueseCompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}
