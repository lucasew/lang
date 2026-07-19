package en

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/en-US/replace.txt
var usReplaceFS embed.FS

var (
	usReplaceOnce sync.Once
	usReplaceBase *rules.AbstractSimpleReplaceRule2
)

func loadUSReplace() *rules.AbstractSimpleReplaceRule2 {
	usReplaceOnce.Do(func() {
		f, err := usReplaceFS.Open("data/en-US/replace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Java AmericanReplaceRule: STYLE, LocaleViolation, crisps → chips example
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                 "EN_US_SIMPLE_REPLACE",
			Description:        "British words easily confused in American English: $match",
			ShortMsg:           "British word",
			MessageTemplate:    "'$match' is a common British expression. Consider using expressions more common to American English.",
			CaseSens:           rules.CaseInsensitive,
			LanguageCode:       "en-US",
			SubRuleSpecificIDs: true,
			Category:           rules.CatStyle.GetCategory(nil),
			IssueType:          rules.ITSLocaleViolation,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/en/en-US/replace.txt"); err != nil {
			panic(err)
		}
		base.AddExamplePair(
			rules.Wrong("Are baked <marker>crisps</marker> healthy?"),
			rules.Fixed("Are baked <marker>chips</marker> healthy?"),
		)
		usReplaceBase = base
	})
	return usReplaceBase
}

// AmericanReplaceRule ports org.languagetool.rules.en.AmericanReplaceRule.
type AmericanReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewAmericanReplaceRule(messages map[string]string) *AmericanReplaceRule {
	base := loadUSReplace()
	r := *base
	r.Messages = messages
	r.Category = rules.CatStyle.GetCategory(messages)
	r.IssueType = rules.ITSLocaleViolation
	return &AmericanReplaceRule{AbstractSimpleReplaceRule2: &r}
}

func (r *AmericanReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
