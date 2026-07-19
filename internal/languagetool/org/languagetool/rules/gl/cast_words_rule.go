package gl

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/spanish.txt
var spanishFS embed.FS

var (
	spanishOnce sync.Once
	spanishMap  map[string][]string
)

func loadSpanish() map[string][]string {
	spanishOnce.Do(func() {
		f, err := spanishFS.Open("data/spanish.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		spanishMap = m
	})
	return spanishMap
}

// CastWordsRule ports org.languagetool.rules.gl.CastWordsRule.
type CastWordsRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewCastWordsRule(messages map[string]string) *CastWordsRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadSpanish(),
		CaseSensitive: false, // Java isCaseSensitive false
		CheckLemmas:   true,  // Java AbstractSimpleReplaceRule default true
		ID:            "GL_CAST_WORDS",
		Description:   "Corrección de erros léxicos (castelanismos).",
		ShortMsg:      "Castelanismos léxicos",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "'" + tokenStr + "' é un castelanismo. Empregue no seu sitio: " +
				joinComma(replacements) + "."
		},
	}
	return &CastWordsRule{AbstractSimpleReplaceRule: base}
}

func (r *CastWordsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
