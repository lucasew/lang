package uk

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace.txt
var replaceFS embed.FS

var (
	replaceOnce sync.Once
	replaceMap  map[string][]string
)

func loadReplace() map[string][]string {
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
		replaceMap = m
	})
	return replaceMap
}

// SimpleReplaceRule ports org.languagetool.rules.uk.SimpleReplaceRule
// (surface match; ignoreTaggedWords / lemma path not modeled without tagger).
type SimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewSimpleReplaceRule(messages map[string]string) *SimpleReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadReplace(),
		CaseSensitive: false,
		CheckLemmas:   false,
		ID:            "UK_SIMPLE_REPLACE",
		Description:   "Пошук помилкових слів",
		ShortMsg:      "Помилка?",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "«" + tokenStr + "» - помилкове слово, виправлення: " + joinCommaUK(replacements) + "."
		},
	}
	return &SimpleReplaceRule{AbstractSimpleReplaceRule: base}
}

func joinCommaUK(ss []string) string {
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
