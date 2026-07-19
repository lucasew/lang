package ar

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replaces.txt
var replacesFS embed.FS

var (
	replacesOnce sync.Once
	replacesBase *rules.AbstractSimpleReplaceRule2
)

func loadReplaces() *rules.AbstractSimpleReplaceRule2 {
	replacesOnce.Do(func() {
		f, err := replacesFS.Open("data/replaces.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "AR_SIMPLE_REPLACE",
			Description:          "قاعدة تطابق الكلمات التي يجب تجنبها وتقترح تصويبا لها",
			ShortMsg:             "خطأ، يفضل أن  يقال:",
			MessageTemplate:      "قل $suggestions",
			SuggestionsSeparator: " أو ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "ar",
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "ar/replaces.txt"); err != nil {
			panic(err)
		}
		// Java: الى → إلى
		base.AddExamplePair(
			rules.Wrong("<marker>الى</marker>"),
			rules.Fixed("<marker>إلى</marker>"),
		)
		replacesBase = base
	})
	return replacesBase
}

// ArabicSimpleReplaceRule ports org.languagetool.rules.ar.ArabicSimpleReplaceRule.
type ArabicSimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewArabicSimpleReplaceRule(messages map[string]string) *ArabicSimpleReplaceRule {
	base := loadReplaces()
	r := *base
	r.Messages = messages
	return &ArabicSimpleReplaceRule{AbstractSimpleReplaceRule2: &r}
}

func (r *ArabicSimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
