package ar

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/wordiness.txt
var wordinessFS embed.FS

var (
	wordinessOnce sync.Once
	wordinessBase *rules.AbstractSimpleReplaceRule2
)

func loadWordiness() *rules.AbstractSimpleReplaceRule2 {
	wordinessOnce.Do(func() {
		f, err := wordinessFS.Open("data/wordiness.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "AR_WORDINESS_REPLACE",
			Description:          "2. حشو(تعبير فيه تكرار)",
			ShortMsg:             "حشو (تعبير فيه تكرار)",
			MessageTemplate:      "'$match' تعبير فيه حشو يفضل أن يقال $suggestions",
			SuggestionsSeparator: " أو ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "ar",
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/ar/wordiness.txt"); err != nil {
			panic(err)
		}
		wordinessBase = base
	})
	return wordinessBase
}

// ArabicWordinessRule ports org.languagetool.rules.ar.ArabicWordinessRule.
type ArabicWordinessRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewArabicWordinessRule(messages map[string]string) *ArabicWordinessRule {
	base := loadWordiness()
	r := *base
	r.Messages = messages
	return &ArabicWordinessRule{AbstractSimpleReplaceRule2: &r}
}

func (r *ArabicWordinessRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
