package es

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace_verbs.txt
var verbsFS embed.FS

var (
	verbsOnce sync.Once
	verbsMap  map[string][]string
)

func loadVerbs() map[string][]string {
	verbsOnce.Do(func() {
		f, err := verbsFS.Open("data/replace_verbs.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		verbsMap = m
	})
	return verbsMap
}

// SimpleReplaceVerbsRule ports org.languagetool.rules.es.SimpleReplaceVerbsRule
// for surface dictionary keys only (no conjugation stripping / synthesizer).
type SimpleReplaceVerbsRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewSimpleReplaceVerbsRule(messages map[string]string) *SimpleReplaceVerbsRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadVerbs(),
		CaseSensitive: false,
		CheckLemmas:   false,
		ID:            "ES_SIMPLE_REPLACE_VERBS",
		Description:   "Detecta verbos incorrectos y propone sugerencias.",
		ShortMsg:      "Verbo incorrecto",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Verbo incorrecto: " + tokenStr
		},
	}
	return &SimpleReplaceVerbsRule{AbstractSimpleReplaceRule: base}
}

func (r *SimpleReplaceVerbsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
