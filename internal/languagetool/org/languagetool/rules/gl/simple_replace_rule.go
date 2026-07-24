package gl

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/words.txt
var wordsFS embed.FS

var (
	wordsOnce  sync.Once
	wordsMap   map[string][]string
)

func loadWords() map[string][]string {
	wordsOnce.Do(func() {
		f, err := wordsFS.Open("data/words.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		wordsMap = m
	})
	return wordsMap
}

// SimpleReplaceRule ports org.languagetool.rules.gl.SimpleReplaceRule.
type SimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewSimpleReplaceRule(messages map[string]string) *SimpleReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadWords(),
		CaseSensitive: false, // Java isCaseSensitive false
		CheckLemmas:   true,  // Java AbstractSimpleReplaceRule default true
		ID:            "GL_SIMPLE_REPLACE",
		Description:   "Corrección de erros léxicos (barbarismos).",
		ShortMsg:      "Erros léxicos",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "'" + tokenStr + "' non existe en galego. Talvez quería vostede dicir: " +
				joinComma(replacements) + "."
		},
	}
	return &SimpleReplaceRule{AbstractSimpleReplaceRule: base}
}

func joinComma(ss []string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += ", "
		}
		out += s
	}
	return out
}

func (r *SimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
