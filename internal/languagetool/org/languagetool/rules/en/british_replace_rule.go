package en

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/en-GB/replace.txt
var gbReplaceFS embed.FS

var (
	gbReplaceOnce sync.Once
	gbReplaceBase *rules.AbstractSimpleReplaceRule2
)

func loadGBReplace() *rules.AbstractSimpleReplaceRule2 {
	gbReplaceOnce.Do(func() {
		f, err := gbReplaceFS.Open("data/en-GB/replace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                 "EN_GB_SIMPLE_REPLACE",
			Description:        "American words easily confused in British English: $match",
			ShortMsg:           "American word",
			MessageTemplate:    "'$match' is a common American expression. Consider using expressions more common to British English.",
			CaseSens:           rules.CaseInsensitive,
			LanguageCode:       "en-GB",
			SubRuleSpecificIDs: true,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/en/en-GB/replace.txt"); err != nil {
			panic(err)
		}
		gbReplaceBase = base
	})
	return gbReplaceBase
}

// BritishReplaceRule ports org.languagetool.rules.en.BritishReplaceRule.
type BritishReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewBritishReplaceRule(messages map[string]string) *BritishReplaceRule {
	base := loadGBReplace()
	r := *base
	r.Messages = messages
	return &BritishReplaceRule{AbstractSimpleReplaceRule2: &r}
}

func (r *BritishReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
