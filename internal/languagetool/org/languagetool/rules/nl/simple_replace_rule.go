package nl

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

func loadReplace() *rules.AbstractSimpleReplaceRule2 {
	replaceOnce.Do(func() {
		f, err := replaceFS.Open("data/replace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "NL_SIMPLE_REPLACE",
			Description:          "Snelle correctie van veel voorkomende vergissingen ($match)",
			ShortMsg:             "Vergissing",
			MessageTemplate:      "Bedoelde u $suggestions?",
			SuggestionsSeparator: " of ",
			// Java: CaseSensitivy.CS (case-sensitive)
			CaseSens:           rules.CaseSensitive,
			LanguageCode:       "nl",
			SubRuleSpecificIDs: true,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/nl/replace.txt"); err != nil {
			panic(err)
		}
		// Java: klaa → klaar
		base.AddExamplePair(
			rules.Wrong("<marker>klaa</marker>."),
			rules.Fixed("<marker>klaar</marker>."),
		)
		replaceBase = base
	})
	return replaceBase
}

// SimpleReplaceRule ports org.languagetool.rules.nl.SimpleReplaceRule.
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
