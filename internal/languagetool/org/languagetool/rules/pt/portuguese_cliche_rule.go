package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/cliches.txt
var clichesFS embed.FS

var (
	clicheOnce sync.Once
	clicheBase *rules.AbstractSimpleReplaceRule2
)

func loadCliche() *rules.AbstractSimpleReplaceRule2 {
	clicheOnce.Do(func() {
		f, err := clichesFS.Open("data/cliches.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "PT_CLICHE_REPLACE",
			Description:          "Frases-feitas e expressões idiomáticas: $match",
			ShortMsg:             "Frase-feita",
			MessageTemplate:      "\"$match\" é uma frase-feita. É preferível dizer $suggestions.",
			SuggestionsSeparator: " ou ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "pt",
			SubRuleSpecificIDs:   true,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/pt/cliches.txt"); err != nil {
			panic(err)
		}
		// Java: quente como uma fornalha → quente
		base.AddExamplePair(
			rules.Wrong("<marker>quente como uma fornalha</marker>"),
			rules.Fixed("<marker>quente</marker>"),
		)
		clicheBase = base
	})
	return clicheBase
}

// PortugueseClicheRule ports org.languagetool.rules.pt.PortugueseClicheRule.
type PortugueseClicheRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewPortugueseClicheRule(messages map[string]string) *PortugueseClicheRule {
	base := loadCliche()
	r := *base
	r.Messages = messages
	return &PortugueseClicheRule{AbstractSimpleReplaceRule2: &r}
}

func (r *PortugueseClicheRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
