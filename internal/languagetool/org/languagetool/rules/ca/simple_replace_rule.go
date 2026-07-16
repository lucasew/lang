package ca

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
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

// SimpleReplaceRule ports org.languagetool.rules.ca.SimpleReplaceRule
// without ConvertToGenderAndNumberFilter (surface suggestions only).
type SimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewSimpleReplaceRule(messages map[string]string) *SimpleReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadReplaceWords(),
		CaseSensitive: false,
		CheckLemmas:   false,
		ID:            "CA_SIMPLE_REPLACE_SIMPLE",
		Description:   "Paraula incorrecta: $match",
		ShortMsg:      "Paraula incorrecta",
		// Stand-in for ignoreTaggedWords: skip capitalized forms that appear as valid
		// proper-name alternatives in the replace list (Navarro, Jerez, …).
		TokenException: func(token *languagetool.AnalyzedTokenReadings) bool {
			t := token.GetToken()
			if !tools.IsCapitalizedWord(t) {
				return false
			}
			for _, rep := range loadReplaceWords()[strings.ToLower(t)] {
				if rep == t {
					return true
				}
			}
			return false
		},
		MessageFn: func(tokenStr string, replacements []string) string {
			if len(replacements) > 0 {
				return "¿Volíeu dir «" + replacements[0] + "»?"
			}
			return "Paraula incorrecta"
		},
	}
	return &SimpleReplaceRule{AbstractSimpleReplaceRule: base}
}

func (r *SimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
