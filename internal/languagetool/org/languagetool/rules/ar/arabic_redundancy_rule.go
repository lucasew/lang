package ar

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
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "AR_REDUNDANCY_REPLACE",
			Description:          "1. تكرار (عام)",
			ShortMsg:             "تكرار",
			MessageTemplate:      "'$match' تعبير فيه تكرار.في بعض الحالات، يستحسن استعمال $suggestions",
			SuggestionsSeparator: " أو ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "ar",
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/ar/redundancies.txt"); err != nil {
			panic(err)
		}
		// Java: سوف لن → لن
		base.AddExamplePair(
			rules.Wrong("<marker>سوف لن</marker>"),
			rules.Fixed("<marker>لن</marker>"),
		)
		redundancyBase = base
	})
	return redundancyBase
}

// ArabicRedundancyRule ports org.languagetool.rules.ar.ArabicRedundancyRule.
type ArabicRedundancyRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewArabicRedundancyRule(messages map[string]string) *ArabicRedundancyRule {
	base := loadRedundancy()
	r := *base
	r.Messages = messages
	return &ArabicRedundancyRule{AbstractSimpleReplaceRule2: &r}
}

func (r *ArabicRedundancyRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
