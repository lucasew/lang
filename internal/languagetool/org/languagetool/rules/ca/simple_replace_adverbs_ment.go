package ca

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace_adverbs_ment.txt
var adverbsMentFS embed.FS

var (
	adverbsMentOnce sync.Once
	adverbsMentMap  map[string][]string
)

func loadAdverbsMent() map[string][]string {
	adverbsMentOnce.Do(func() {
		f, err := adverbsMentFS.Open("data/replace_adverbs_ment.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		adverbsMentMap = m
	})
	return adverbsMentMap
}

// SimpleReplaceAdverbsMent ports org.languagetool.rules.ca.SimpleReplaceAdverbsMent.
type SimpleReplaceAdverbsMent struct {
	*rules.AbstractSimpleReplaceRule
}

func NewSimpleReplaceAdverbsMent(messages map[string]string) *SimpleReplaceAdverbsMent {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadAdverbsMent(),
		CaseSensitive: false,
		CheckLemmas:   false,
		ID:            "ADVERBIS_MENT",
		Description:   "Alternatives a adverbis acabats en -ment: $match",
		ShortMsg:      "Alternatives a adverbis acabats en -ment",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "A vegades s'abusa dels adverbis acabats en -ment en detriment de formes més àgils."
		},
	}
	return &SimpleReplaceAdverbsMent{AbstractSimpleReplaceRule: base}
}

func (r *SimpleReplaceAdverbsMent) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
