package ar

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/darja.txt
var darjaFS embed.FS

var (
	darjaOnce sync.Once
	darjaBase *rules.AbstractSimpleReplaceRule2
)

func loadDarja() *rules.AbstractSimpleReplaceRule2 {
	darjaOnce.Do(func() {
		f, err := darjaFS.Open("data/darja.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "AR_DARJA_REPLACE",
			Description:          "كلمات بديلة للكلمات العامية أو الأجنبية",
			ShortMsg:             "كلمات بديلة للكلمات العامية أو الأجنبية",
			MessageTemplate:      "الكلمة عامية  أو أجنبية يفضل أن يقال $suggestions",
			SuggestionsSeparator: " أو  ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "ar",
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/ar/darja.txt"); err != nil {
			panic(err)
		}
		// Java: طرشي → فلفل حلو
		base.AddExamplePair(
			rules.Wrong("<marker>طرشي</marker>"),
			rules.Fixed("<marker>فلفل حلو</marker>"),
		)
		darjaBase = base
	})
	return darjaBase
}

// ArabicDarjaRule ports org.languagetool.rules.ar.ArabicDarjaRule.
type ArabicDarjaRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewArabicDarjaRule(messages map[string]string) *ArabicDarjaRule {
	base := loadDarja()
	r := *base
	r.Messages = messages
	return &ArabicDarjaRule{AbstractSimpleReplaceRule2: &r}
}

func (r *ArabicDarjaRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
