package es

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/wikipedia.txt
var wikipediaFS embed.FS

var (
	wikipediaOnce sync.Once
	wikipediaBase *rules.AbstractSimpleReplaceRule2
)

func loadWikipedia() *rules.AbstractSimpleReplaceRule2 {
	wikipediaOnce.Do(func() {
		f, err := wikipediaFS.Open("data/wikipedia.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "ES_WIKIPEDIA_COMMON_ERRORS",
			Description:          "Errores frecuentes en los artículos de la Wikipedia",
			ShortMsg:             "Error gramatical u ortográfico",
			MessageTemplate:      "'$match' es una expresión errónea. Pruebe a utilizar $suggestions",
			SuggestionsSeparator: " o ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "es",
			SubRuleSpecificIDs:   true,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/es/wikipedia.txt"); err != nil {
			panic(err)
		}
		wikipediaBase = base
	})
	return wikipediaBase
}

// SpanishWikipediaRule ports org.languagetool.rules.es.SpanishWikipediaRule.
type SpanishWikipediaRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewSpanishWikipediaRule(messages map[string]string) *SpanishWikipediaRule {
	base := loadWikipedia()
	r := *base
	r.Messages = messages
	return &SpanishWikipediaRule{AbstractSimpleReplaceRule2: &r}
}

func (r *SpanishWikipediaRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
