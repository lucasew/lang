package jekavian

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

// SimpleGrammarJekavianReplaceRule ports org.languagetool.rules.sr.jekavian.SimpleGrammarJekavianReplaceRule.
type SimpleGrammarJekavianReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewSimpleGrammarJekavianReplaceRule(messages map[string]string) *SimpleGrammarJekavianReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadGrammar(),
		CaseSensitive: false,
		CheckLemmas:   true, // Java default checkLemmas true
		ID:            "SR_JEKAVIAN_SIMPLE_GRAMMAR_REPLACE_RULE",
		Description:   "Провера граматички погрешних ријечи или израза",
		ShortMsg:      "Граматички погрешна ријеч тј. израз",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Не каже се „" + tokenStr + "“ него „" + joinComma(replacements) + "“."
		},
	}
	return &SimpleGrammarJekavianReplaceRule{AbstractSimpleReplaceRule: base}
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

func (r *SimpleGrammarJekavianReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
