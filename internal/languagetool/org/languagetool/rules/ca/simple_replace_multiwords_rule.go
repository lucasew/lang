package ca

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace_multiwords.txt
var multiwordsFS embed.FS

var (
	multiwordsOnce sync.Once
	multiwordsBase *rules.AbstractSimpleReplaceRule2
)

func loadMultiwords() *rules.AbstractSimpleReplaceRule2 {
	multiwordsOnce.Do(func() {
		f, err := multiwordsFS.Open("data/replace_multiwords.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "CA_SIMPLE_REPLACE_MULTIWORDS",
			Description:          "Expressions inadequades: $match",
			ShortMsg:             "Expressió inadequada",
			MessageTemplate:      "Expressió incorrecta.",
			SuggestionsSeparator: " o ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "ca",
			SubRuleSpecificIDs:   true,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/ca/replace_multiwords.txt"); err != nil {
			panic(err)
		}
		multiwordsBase = base
	})
	return multiwordsBase
}

// SimpleReplaceMultiwordsRule ports org.languagetool.rules.ca.SimpleReplaceMultiwordsRule.
type SimpleReplaceMultiwordsRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewSimpleReplaceMultiwordsRule(messages map[string]string) *SimpleReplaceMultiwordsRule {
	base := loadMultiwords()
	r := *base
	r.Messages = messages
	return &SimpleReplaceMultiwordsRule{AbstractSimpleReplaceRule2: &r}
}

func (r *SimpleReplaceMultiwordsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
