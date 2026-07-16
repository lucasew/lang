package en

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace.txt data/replace_custom.txt
var enReplaceFS embed.FS

var (
	enReplaceOnce sync.Once
	enReplaceBase *rules.AbstractSimpleReplaceRule2
)

func loadENReplace() *rules.AbstractSimpleReplaceRule2 {
	enReplaceOnce.Do(func() {
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "EN_SIMPLE_REPLACE",
			Description:          "Check for wrong words/phrases: $match",
			ShortMsg:             "Wrong word",
			MessageTemplate:      "Did you mean $suggestions?",
			SuggestionsSeparator: " or ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "en",
			SubRuleSpecificIDs:   true,
		}
		for _, name := range []string{"data/replace.txt", "data/replace_custom.txt"} {
			f, err := enReplaceFS.Open(name)
			if err != nil {
				continue
			}
			if err := base.LoadSimpleReplaceRule2Data(f, "/en/"+name); err != nil {
				f.Close()
				panic(err)
			}
			f.Close()
		}
		enReplaceBase = base
	})
	return enReplaceBase
}

// SimpleReplaceRule ports org.languagetool.rules.en.SimpleReplaceRule.
type SimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewSimpleReplaceRule(messages map[string]string) *SimpleReplaceRule {
	base := loadENReplace()
	r := *base
	r.Messages = messages
	return &SimpleReplaceRule{AbstractSimpleReplaceRule2: &r}
}

func (r *SimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
