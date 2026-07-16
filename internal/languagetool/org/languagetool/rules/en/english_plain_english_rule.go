package en

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/wordiness.txt
var wordinessFS embed.FS

var (
	plainOnce sync.Once
	plainBase *rules.AbstractSimpleReplaceRule2
)

func loadPlainEnglish() *rules.AbstractSimpleReplaceRule2 {
	plainOnce.Do(func() {
		f, err := wordinessFS.Open("data/wordiness.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "EN_PLAIN_ENGLISH_REPLACE",
			Description:          "1. Wordiness (General)",
			ShortMsg:             "Wordiness",
			MessageTemplate:      "'$match' is a wordy or complex expression. In some cases, it might be preferable to use $suggestions.",
			SuggestionsSeparator: " or ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "en",
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/en/wordiness.txt"); err != nil {
			panic(err)
		}
		plainBase = base
	})
	return plainBase
}

// EnglishPlainEnglishRule ports org.languagetool.rules.en.EnglishPlainEnglishRule.
type EnglishPlainEnglishRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewEnglishPlainEnglishRule(messages map[string]string) *EnglishPlainEnglishRule {
	base := loadPlainEnglish()
	r := *base
	r.Messages = messages
	return &EnglishPlainEnglishRule{AbstractSimpleReplaceRule2: &r}
}

func (r *EnglishPlainEnglishRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
