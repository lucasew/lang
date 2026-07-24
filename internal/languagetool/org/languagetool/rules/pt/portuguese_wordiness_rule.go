package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/wordiness.txt
var wordinessFS embed.FS

var (
	wordinessOnce sync.Once
	wordinessBase *rules.AbstractSimpleReplaceRule2
)

func loadWordiness() *rules.AbstractSimpleReplaceRule2 {
	wordinessOnce.Do(func() {
		f, err := wordinessFS.Open("data/wordiness.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "PT_WORDINESS_REPLACE",
			Description:          "2. Expressões prolixas: $match",
			ShortMsg:             "Expressão prolixa",
			MessageTemplate:      "\"$match\" é uma expressão prolixa. É preferível dizer $suggestions.",
			SuggestionsSeparator: " ou ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "pt",
			SubRuleSpecificIDs:   true,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/pt/wordiness.txt"); err != nil {
			panic(err)
		}
		// Java: Raramente é o caso em que acontece → Raramente acontece
		base.AddExamplePair(
			rules.Wrong("<marker>Raramente é o caso em que acontece</marker> isto."),
			rules.Fixed("<marker>Raramente acontece</marker> isto."),
		)
		wordinessBase = base
	})
	return wordinessBase
}

// PortugueseWordinessRule ports org.languagetool.rules.pt.PortugueseWordinessRule.
type PortugueseWordinessRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewPortugueseWordinessRule(messages map[string]string) *PortugueseWordinessRule {
	base := loadWordiness()
	r := *base
	r.Messages = messages
	return &PortugueseWordinessRule{AbstractSimpleReplaceRule2: &r}
}

func (r *PortugueseWordinessRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
