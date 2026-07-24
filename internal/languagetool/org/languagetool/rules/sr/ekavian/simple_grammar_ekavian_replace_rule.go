package ekavian

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace-grammar.txt
var grammarFS embed.FS

var (
	grammarOnce sync.Once
	grammarMap  map[string][]string
)

func loadGrammar() map[string][]string {
	grammarOnce.Do(func() {
		f, err := grammarFS.Open("data/replace-grammar.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		grammarMap = m
	})
	return grammarMap
}

// SimpleGrammarEkavianReplaceRule ports org.languagetool.rules.sr.ekavian.SimpleGrammarEkavianReplaceRule.
type SimpleGrammarEkavianReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewSimpleGrammarEkavianReplaceRule(messages map[string]string) *SimpleGrammarEkavianReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadGrammar(),
		CaseSensitive: false,
		CheckLemmas:   true, // Java default checkLemmas true
		ID:            "SR_EKAVIAN_SIMPLE_GRAMMAR_REPLACE_RULE",
		Description:   "Провера граматички погрешних речи или израза",
		ShortMsg:      "Граматички погрешна реч тј. израз",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Уместо «" + tokenStr + "» треба рећи: " + joinComma(replacements) + "."
		},
	}
	return &SimpleGrammarEkavianReplaceRule{AbstractSimpleReplaceRule: base}
}

func (r *SimpleGrammarEkavianReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
