package de

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace.txt data/replace_custom.txt
var replaceFS embed.FS

var (
	replaceOnce sync.Once
	replaceBase *rules.AbstractSimpleReplaceRule2
)

func loadReplace() *rules.AbstractSimpleReplaceRule2 {
	replaceOnce.Do(func() {
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "DE_SIMPLE_REPLACE",
			Description:          "Prüft auf bestimmte falsche Wörter/Phrasen: $match",
			ShortMsg:             "Falsches Wort",
			MessageTemplate:      "Meinten Sie vielleicht $suggestions?",
			SuggestionsSeparator: " oder ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "de",
			SubRuleSpecificIDs:   true,
		}
		for _, name := range []string{"data/replace.txt", "data/replace_custom.txt"} {
			f, err := replaceFS.Open(name)
			if err != nil {
				continue
			}
			if err := base.LoadSimpleReplaceRule2Data(f, "/de/"+name); err != nil {
				f.Close()
				panic(err)
			}
			f.Close()
		}
		replaceBase = base
	})
	return replaceBase
}

// SimpleReplaceRule ports org.languagetool.rules.de.SimpleReplaceRule.
type SimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewSimpleReplaceRule(messages map[string]string) *SimpleReplaceRule {
	base := loadReplace()
	r := *base
	r.Messages = messages
	return &SimpleReplaceRule{AbstractSimpleReplaceRule2: &r}
}

func (r *SimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
