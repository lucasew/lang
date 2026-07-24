package pt

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
			ID:                   "PT_WIKIPEDIA_COMMON_ERRORS",
			Description:          "Erros frequentes nos artigos da Wikipédia: $match",
			ShortMsg:             "Erro gramatical ou de normativa",
			MessageTemplate:      "Possível erro em \"$match\". Prefira $suggestions",
			SuggestionsSeparator: " ou ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "pt",
			SubRuleSpecificIDs:   true,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/pt/wikipedia.txt"); err != nil {
			panic(err)
		}
		wikipediaBase = base
	})
	return wikipediaBase
}

// PortugueseWikipediaRule ports org.languagetool.rules.pt.PortugueseWikipediaRule.
type PortugueseWikipediaRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewPortugueseWikipediaRule(messages map[string]string) *PortugueseWikipediaRule {
	base := loadWikipedia()
	r := *base
	r.Messages = messages
	return &PortugueseWikipediaRule{AbstractSimpleReplaceRule2: &r}
}

func (r *PortugueseWikipediaRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
