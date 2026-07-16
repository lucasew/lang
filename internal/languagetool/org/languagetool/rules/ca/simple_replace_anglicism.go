package ca

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace_anglicism.txt
var anglicismFS embed.FS

var (
	anglicismOnce sync.Once
	anglicismBase *rules.AbstractSimpleReplaceRule2
)

func loadAnglicism() *rules.AbstractSimpleReplaceRule2 {
	anglicismOnce.Do(func() {
		f, err := anglicismFS.Open("data/replace_anglicism.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "CA_SIMPLE_REPLACE_ANGLICISM",
			Description:          "Anglicismes innecessaris: $match",
			ShortMsg:             "Anglicisme innecessari",
			MessageTemplate:      "Anglicisme innecessari. Considereu fer servir una altra paraula.",
			SuggestionsSeparator: " o ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "ca",
			SubRuleSpecificIDs:   true,
			// Without ConvertToGenderAndNumberFilter and NP/_english_ignore_ tags.
			IsTokenException: func(atr *languagetool.AnalyzedTokenReadings) bool {
				return atr.IsImmunized()
			},
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/ca/replace_anglicism.txt"); err != nil {
			panic(err)
		}
		anglicismBase = base
	})
	return anglicismBase
}

// SimpleReplaceAnglicism ports org.languagetool.rules.ca.SimpleReplaceAnglicism
// without ConvertToGenderAndNumberFilter (surface suggestions only).
type SimpleReplaceAnglicism struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewSimpleReplaceAnglicism(messages map[string]string) *SimpleReplaceAnglicism {
	base := loadAnglicism()
	r := *base
	r.Messages = messages
	return &SimpleReplaceAnglicism{AbstractSimpleReplaceRule2: &r}
}

func (r *SimpleReplaceAnglicism) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
