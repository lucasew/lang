package en

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/diacritics.txt
var diacriticsFS embed.FS

var (
	diacriticsOnce sync.Once
	diacriticsBase *rules.AbstractSimpleReplaceRule2
)

func loadDiacritics() *rules.AbstractSimpleReplaceRule2 {
	diacriticsOnce.Do(func() {
		f, err := diacriticsFS.Open("data/diacritics.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "EN_DIACRITICS_REPLACE",
			Description:          "Suggest diacritics for '$match'",
			ShortMsg:             "The original word has a diacritic",
			MessageTemplate:      "'$match' is an imported foreign name or expression, which originally has a diacritic.",
			SuggestionsSeparator: " or ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "en",
			SubRuleSpecificIDs:   true,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/en/diacritics.txt"); err != nil {
			panic(err)
		}
		diacriticsBase = base
	})
	return diacriticsBase
}

// EnglishDiacriticsRule ports org.languagetool.rules.en.EnglishDiacriticsRule.
type EnglishDiacriticsRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewEnglishDiacriticsRule(messages map[string]string) *EnglishDiacriticsRule {
	base := loadDiacritics()
	r := *base
	r.Messages = messages
	return &EnglishDiacriticsRule{AbstractSimpleReplaceRule2: &r}
}

func (r *EnglishDiacriticsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
