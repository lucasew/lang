package fr

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace.txt data/replace_custom.txt
var replaceFS embed.FS

var (
	replaceOnce  sync.Once
	replaceWords map[string][]string
)

func loadReplaceWords() map[string][]string {
	replaceOnce.Do(func() {
		m := map[string][]string{}
		for _, name := range []string{"data/replace.txt", "data/replace_custom.txt"} {
			f, err := replaceFS.Open(name)
			if err != nil {
				continue
			}
			part, err := rules.LoadSimpleReplaceWords(f)
			f.Close()
			if err != nil {
				panic(err)
			}
			for k, v := range part {
				m[k] = v
			}
		}
		replaceWords = m
	})
	return replaceWords
}

// SimpleReplaceRule ports org.languagetool.rules.fr.SimpleReplaceRule.
type SimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewSimpleReplaceRule(messages map[string]string) *SimpleReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadReplaceWords(),
		CaseSensitive: false,
		CheckLemmas:   false,
		ID:            "FR_SIMPLE_REPLACE_SIMPLE",
		LanguageCode:         "fr",
		SubRuleSpecificIDs:   true,
		Description:   "Mot incorrect : $match",
		ShortMsg:      "Mot incorrect",
		MessageFn: func(tokenStr string, replacements []string) string {
			if len(replacements) > 0 {
				return "Vouliez-vous dire « " + replacements[0] + " » ?"
			}
			return "Mot incorrect"
		},
	}
	return &SimpleReplaceRule{AbstractSimpleReplaceRule: base}
}

func (r *SimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
