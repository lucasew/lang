package en

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/redundancies.txt
var redundancyFS embed.FS

var (
	redundancyOnce sync.Once
	redundancyBase *rules.AbstractSimpleReplaceRule2
)

func loadRedundancy() *rules.AbstractSimpleReplaceRule2 {
	redundancyOnce.Do(func() {
		f, err := redundancyFS.Open("data/redundancies.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Java EnglishRedundancyRule: REDUNDANCY, Style, setDefaultOff, Wikipedia URL, tuna fish → tuna
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "EN_REDUNDANCY_REPLACE",
			Description:          "1. Redundancy (General)",
			ShortMsg:             "Redundancy",
			MessageTemplate:      "'$match' is a redundancy. In some cases, it might be preferable to use $suggestions",
			SuggestionsSeparator: " or ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "en",
			SubRuleSpecificIDs:   true,
			Category:             rules.CatRedundancy.GetCategory(nil),
			IssueType:            rules.ITSStyle,
			DefaultOff:           true,
			URL:                  "https://en.wikipedia.org/wiki/Redundancy_(linguistics)",
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/en/redundancies.txt"); err != nil {
			panic(err)
		}
		base.AddExamplePair(
			rules.Wrong("<marker>tuna fish</marker>"),
			rules.Fixed("<marker>tuna</marker>"),
		)
		redundancyBase = base
	})
	return redundancyBase
}

// EnglishRedundancyRule ports org.languagetool.rules.en.EnglishRedundancyRule.
type EnglishRedundancyRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewEnglishRedundancyRule(messages map[string]string) *EnglishRedundancyRule {
	base := loadRedundancy()
	r := *base
	r.Messages = messages
	r.Category = rules.CatRedundancy.GetCategory(messages)
	r.IssueType = rules.ITSStyle
	r.DefaultOff = true
	r.URL = "https://en.wikipedia.org/wiki/Redundancy_(linguistics)"
	return &EnglishRedundancyRule{AbstractSimpleReplaceRule2: &r}
}

func (r *EnglishRedundancyRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
