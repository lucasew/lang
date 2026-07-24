package km

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/coherency.txt
var replaceFS embed.FS

var (
	replaceOnce sync.Once
	replaceBase *rules.AbstractSimpleReplaceRule2
)

func loadReplace() *rules.AbstractSimpleReplaceRule2 {
	replaceOnce.Do(func() {
		f, err := replaceFS.Open("data/coherency.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "KM_SIMPLE_REPLACE",
			Description:          "Words or groups of words that are incorrect or obsolete",
			ShortMsg:             "Consider following the spelling of Chuon Nath",
			MessageTemplate:      " Consider following the spelling of Chuon Nath ",
			SuggestionsSeparator: " or ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "km",
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/km/coherency.txt"); err != nil {
			panic(err)
		}
		replaceBase = base
	})
	return replaceBase
}

// KhmerSimpleReplaceRule ports org.languagetool.rules.km.KhmerSimpleReplaceRule.
type KhmerSimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewKhmerSimpleReplaceRule(messages map[string]string) *KhmerSimpleReplaceRule {
	base := loadReplace()
	r := *base
	r.Messages = messages
	return &KhmerSimpleReplaceRule{AbstractSimpleReplaceRule2: &r}
}

func (r *KhmerSimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
