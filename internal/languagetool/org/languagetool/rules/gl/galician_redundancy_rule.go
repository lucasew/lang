package gl

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
			ID:                   "GL_REDUNDANCY_REPLACE",
			Description:          "1. Pleonasmos e redundancias",
			ShortMsg:             "Pleonasmo",
			MessageTemplate:      "'$match' é un pleonasmo. É preferible dicir $suggestions",
			SuggestionsSeparator: " ou ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "gl",
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/gl/redundancies.txt"); err != nil {
			panic(err)
		}
		redundancyBase = base
	})
	return redundancyBase
}

// GalicianRedundancyRule ports org.languagetool.rules.gl.GalicianRedundancyRule.
type GalicianRedundancyRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewGalicianRedundancyRule(messages map[string]string) *GalicianRedundancyRule {
	base := loadRedundancy()
	r := *base
	r.Messages = messages
	return &GalicianRedundancyRule{AbstractSimpleReplaceRule2: &r}
}

func (r *GalicianRedundancyRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
