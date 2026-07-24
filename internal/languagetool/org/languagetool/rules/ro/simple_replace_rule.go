package ro

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
	replaceBase *rules.AbstractSimpleReplaceRule2
)

func loadReplaceRule() *rules.AbstractSimpleReplaceRule2 {
	replaceOnce.Do(func() {
		f, err := replaceFS.Open("data/replace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "RO_SIMPLE_REPLACE",
			Description:          "Cuvinte sau grupuri de cuvinte incorecte sau ieșite din uz",
			ShortMsg:             "Cuvânt incorect sau ieșit din uz",
			MessageTemplate:      "'$match' este incorect sau ieșit din uz, folosiți $suggestions",
			SuggestionsSeparator: " sau ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "ro",
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/ro/replace.txt"); err != nil {
			panic(err)
		}
		replaceBase = base
	})
	return replaceBase
}

// SimpleReplaceRule ports org.languagetool.rules.ro.SimpleReplaceRule.
type SimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewSimpleReplaceRule(messages map[string]string) *SimpleReplaceRule {
	base := loadReplaceRule()
	r := *base
	r.Messages = messages
	return &SimpleReplaceRule{AbstractSimpleReplaceRule2: &r}
}

func (r *SimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
