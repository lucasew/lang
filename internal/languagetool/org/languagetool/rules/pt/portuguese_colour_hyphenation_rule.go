package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/compound_colours.txt
var colourCompoundFS embed.FS

var (
	colourCompoundOnce sync.Once
	colourCompoundData *rules.CompoundRuleData
)

func loadColourCompounds() *rules.CompoundRuleData {
	colourCompoundOnce.Do(func() {
		f, err := colourCompoundFS.Open("data/compound_colours.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		d, err := rules.NewCompoundRuleData(f, "/pt/compound_colours.txt")
		if err != nil {
			panic(err)
		}
		colourCompoundData = d
	})
	return colourCompoundData
}

// PortugueseColourHyphenationRule ports org.languagetool.rules.pt.PortugueseColourHyphenationRule.
type PortugueseColourHyphenationRule struct {
	*rules.AbstractCompoundRule
}

func NewPortugueseColourHyphenationRule(messages map[string]string) *PortugueseColourHyphenationRule {
	base := &rules.AbstractCompoundRule{
		Messages:                   messages,
		ID:                         "PT_COLOUR_HYPHENATION",
		Description:                "Nomes de cores devem ser hifenizados: \"$match\"",
		WithHyphenMessage:          "Nomes de cores são palavras compostas e devem ser hifenizados.",
		WithoutHyphenMessage:       "Esta palavra é composta por justaposição.",
		WithOrWithoutHyphenMessage: "Esta palavra pode ser composta por justaposição ou hifenizada.",
		ShortDesc:                  "Nomes de cores são palavras compostas.",
		Data:                       loadColourCompounds(),
	}
	base.UseSubRuleSpecificIDs()
	return &PortugueseColourHyphenationRule{AbstractCompoundRule: base}
}

func (r *PortugueseColourHyphenationRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}
