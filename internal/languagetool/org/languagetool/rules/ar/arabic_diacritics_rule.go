package ar

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
			ID:                   "AR_DIACRITICS_REPLACE",
			Description:          "كلمات مشكولة للتوضيح",
			ShortMsg:             "كلمات يستحسن أن تشكّل لتصحيح نطقها",
			MessageTemplate:      "'$match' كلمة يشيع نطقها نطقا خاطئا لذا نقترح تشكيلها كالآتي: $suggestions",
			SuggestionsSeparator: " أو ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "ar",
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/ar/diacritics.txt"); err != nil {
			panic(err)
		}
		diacriticsBase = base
	})
	return diacriticsBase
}

// ArabicDiacriticsRule ports org.languagetool.rules.ar.ArabicDiacriticsRule.
type ArabicDiacriticsRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewArabicDiacriticsRule(messages map[string]string) *ArabicDiacriticsRule {
	base := loadDiacritics()
	r := *base
	r.Messages = messages
	return &ArabicDiacriticsRule{AbstractSimpleReplaceRule2: &r}
}

func (r *ArabicDiacriticsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
