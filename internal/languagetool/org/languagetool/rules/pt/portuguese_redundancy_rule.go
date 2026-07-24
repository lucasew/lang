package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/redundancies.txt
var redundancyFS embed.FS

var (
	redundancyOnce sync.Once
	redundancyBase *rules.AbstractSimpleReplaceRule2
)

func loadRedundancy() *rules.AbstractSimpleReplaceRule2 {
	redundancyOnce.Do(func() {
		f, err := redundancyFS.Open("data/redundancies.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "PT_REDUNDANCY_REPLACE",
			Description:          "1. Pleonasmos e redundâncias: $match",
			ShortMsg:             "Pleonasmo",
			MessageTemplate:      "\"$match\" é um pleonasmo. É preferível dizer $suggestions",
			SuggestionsSeparator: " ou ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "pt",
			SubRuleSpecificIDs:   true,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/pt/redundancies.txt"); err != nil {
			panic(err)
		}
		// Java: duna de areia → duna
		base.AddExamplePair(
			rules.Wrong("<marker>duna de areia</marker>"),
			rules.Fixed("<marker>duna</marker>"),
		)
		redundancyBase = base
	})
	return redundancyBase
}

// PortugueseRedundancyRule ports org.languagetool.rules.pt.PortugueseRedundancyRule.
type PortugueseRedundancyRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewPortugueseRedundancyRule(messages map[string]string) *PortugueseRedundancyRule {
	base := loadRedundancy()
	r := *base
	r.Messages = messages
	return &PortugueseRedundancyRule{AbstractSimpleReplaceRule2: &r}
}

func (r *PortugueseRedundancyRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
