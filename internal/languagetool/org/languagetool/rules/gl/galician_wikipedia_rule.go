package gl

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
			ID:                   "GL_WIKIPEDIA_COMMON_ERRORS",
			Description:          "Erros frecuentes nos artigos da Wikipedia",
			ShortMsg:             "Erro gramatical ou de normativa",
			MessageTemplate:      "'$match' é un erro. Considere utilizar $suggestions",
			SuggestionsSeparator: " ou ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "gl",
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/gl/wikipedia.txt"); err != nil {
			panic(err)
		}
		// Java: a efectos de → para os efectos de
		base.AddExamplePair(
			rules.Wrong("<marker>a efectos de</marker>"),
			rules.Fixed("<marker>para os efectos de</marker>"),
		)
		wikipediaBase = base
	})
	return wikipediaBase
}

// GalicianWikipediaRule ports org.languagetool.rules.gl.GalicianWikipediaRule.
type GalicianWikipediaRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewGalicianWikipediaRule(messages map[string]string) *GalicianWikipediaRule {
	base := loadWikipedia()
	r := *base
	r.Messages = messages
	return &GalicianWikipediaRule{AbstractSimpleReplaceRule2: &r}
}

func (r *GalicianWikipediaRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
