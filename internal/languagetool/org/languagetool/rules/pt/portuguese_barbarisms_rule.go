package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/pt-BR/barbarisms.txt
var barbarismsBRFS embed.FS

var (
	barbarismsBROnce sync.Once
	barbarismsBRBase *rules.AbstractSimpleReplaceRule2
)

func loadBarbarismsBR() *rules.AbstractSimpleReplaceRule2 {
	barbarismsBROnce.Do(func() {
		f, err := barbarismsBRFS.Open("data/pt-BR/barbarisms.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "PT_BARBARISMS_REPLACE",
			Description:          "Palavras de origem estrangeira evitáveis: $match",
			ShortMsg:             "Estrangeirismo",
			MessageTemplate:      "\"$match\" é um estrangeirismo. É preferível dizer $suggestions.",
			SuggestionsSeparator: " ou ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "pt",
			SubRuleSpecificIDs:   true,
			// Java also skips NP / _english_ignore_ tags; surface-only port has no POS.
			IsTokenException: func(atr *languagetool.AnalyzedTokenReadings) bool {
				return atr.IsImmunized()
			},
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/pt/pt-BR/barbarisms.txt"); err != nil {
			panic(err)
		}
		barbarismsBRBase = base
	})
	return barbarismsBRBase
}

// PortugueseBarbarismsRule ports org.languagetool.rules.pt.PortugueseBarbarismsRule
// using the pt-BR barbarisms path (as in PortugueseBarbarismRuleTest).
type PortugueseBarbarismsRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewPortugueseBarbarismsRule(messages map[string]string) *PortugueseBarbarismsRule {
	base := loadBarbarismsBR()
	r := *base
	r.Messages = messages
	return &PortugueseBarbarismsRule{AbstractSimpleReplaceRule2: &r}
}

func (r *PortugueseBarbarismsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
