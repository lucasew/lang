package ar

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/homophones.txt
var homophonesFS embed.FS

var (
	homophonesOnce sync.Once
	homophonesBase *rules.AbstractSimpleReplaceRule2
)

func loadHomophones() *rules.AbstractSimpleReplaceRule2 {
	homophonesOnce.Do(func() {
		f, err := homophonesFS.Open("data/homophones.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "AR_HOMOPHONES_REPLACE",
			Description:          "كلمات متشابهة لفظا للتوضيح، يرجى التحقق منها مثل تشابه الظاء والضاد.",
			ShortMsg:             "كلمات متشابهة لفظا يرجى التحقق منها",
			MessageTemplate:      "قل $suggestions",
			SuggestionsSeparator: " أو ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "ar",
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "ar/homophones.txt"); err != nil {
			panic(err)
		}
		homophonesBase = base
	})
	return homophonesBase
}

// ArabicHomophonesRule ports org.languagetool.rules.ar.ArabicHomophonesRule.
type ArabicHomophonesRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewArabicHomophonesRule(messages map[string]string) *ArabicHomophonesRule {
	base := loadHomophones()
	r := *base
	r.Messages = messages
	return &ArabicHomophonesRule{AbstractSimpleReplaceRule2: &r}
}

func (r *ArabicHomophonesRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
