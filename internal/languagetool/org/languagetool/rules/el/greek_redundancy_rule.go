package el

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
			ID:                   "EL_REDUNDANCY_REPLACE",
			Description:          "Έλεγχος για χρήση πλεονασμού σε μια πρόταση.",
			ShortMsg:             "Πλεονασμός",
			MessageTemplate:      "'$match' είναι πλεονασμός. Γενικά, είναι προτιμότερο το: $suggestions",
			SuggestionsSeparator: " ή ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "el",
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/el/redundancies.txt"); err != nil {
			panic(err)
		}
		redundancyBase = base
	})
	return redundancyBase
}

// GreekRedundancyRule ports org.languagetool.rules.el.GreekRedundancyRule.
type GreekRedundancyRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewGreekRedundancyRule(messages map[string]string) *GreekRedundancyRule {
	base := loadRedundancy()
	r := *base
	r.Messages = messages
	return &GreekRedundancyRule{AbstractSimpleReplaceRule2: &r}
}

func (r *GreekRedundancyRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
