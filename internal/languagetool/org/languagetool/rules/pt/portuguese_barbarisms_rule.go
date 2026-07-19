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
			// Java isTokenException: isImmunized || NP* || _english_ignore_*
			IsTokenException: barbarismsTokenException,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/pt/pt-BR/barbarisms.txt"); err != nil {
			panic(err)
		}
		// Java: curriculum vitae → currículo
		base.AddExamplePair(
			rules.Wrong("<marker>curriculum vitae</marker>"),
			rules.Fixed("<marker>currículo</marker>"),
		)
		barbarismsBRBase = base
	})
	return barbarismsBRBase
}

// barbarismsTokenException ports PortugueseBarbarismsRule.isTokenException.
// Without NP / _english_ignore_ tags, only immunization applies (fail closed).
func barbarismsTokenException(atr *languagetool.AnalyzedTokenReadings) bool {
	if atr == nil {
		return false
	}
	if atr.IsImmunized() {
		return true
	}
	if atr.HasPosTagStartingWith("NP") {
		return true
	}
	if atr.HasPosTagStartingWith("_english_ignore_") {
		return true
	}
	return false
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
