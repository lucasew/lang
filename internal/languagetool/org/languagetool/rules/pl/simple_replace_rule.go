package pl

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace.txt
var replaceFS embed.FS

var (
	replaceOnce  sync.Once
	replaceWords map[string][]string
)

func loadReplaceWords() map[string][]string {
	replaceOnce.Do(func() {
		f, err := replaceFS.Open("data/replace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		replaceWords = m
	})
	return replaceWords
}

// SimpleReplaceRule ports org.languagetool.rules.pl.SimpleReplaceRule.
type SimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewSimpleReplaceRule(messages map[string]string) *SimpleReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadReplaceWords(),
		CaseSensitive: false,
		CheckLemmas:   false,
		ID:            "PL_SIMPLE_REPLACE",
		Description:   "Typowe literówki i niepoprawne wyrazy (domowi, sie, niewiadomo, duh, cie…)",
		ShortMsg:      "Literówka",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Wyraz „" + tokenStr + "” to najczęściej literówka; poprawnie pisze się: " +
				strings.Join(replacements, ", ") + "."
		},
	}
	// Java: sei → się
	base.AddExamplePair(
		rules.Wrong("Uspokój <marker>sei</marker>."),
		rules.Fixed("Uspokój <marker>się</marker>."),
	)
	return &SimpleReplaceRule{AbstractSimpleReplaceRule: base}
}

func (r *SimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
