package gl

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/barbarisms.txt
var barbarismsFS embed.FS

var (
	barbarismsOnce sync.Once
	barbarismsBase *rules.AbstractSimpleReplaceRule2
)

func loadBarbarisms() *rules.AbstractSimpleReplaceRule2 {
	barbarismsOnce.Do(func() {
		f, err := barbarismsFS.Open("data/barbarisms.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "GL_BARBARISM_REPLACE",
			Description:          "Palabras de orixe estranxeira evitábeis",
			ShortMsg:             "Xenismo",
			MessageTemplate:      "'$match' é un xenismo. É preferíbel dicir $suggestions",
			SuggestionsSeparator: " ou ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "gl",
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/gl/barbarisms.txt"); err != nil {
			panic(err)
		}
		barbarismsBase = base
	})
	return barbarismsBase
}

// GalicianBarbarismsRule ports org.languagetool.rules.gl.GalicianBarbarismsRule.
type GalicianBarbarismsRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewGalicianBarbarismsRule(messages map[string]string) *GalicianBarbarismsRule {
	base := loadBarbarisms()
	r := *base
	r.Messages = messages
	return &GalicianBarbarismsRule{AbstractSimpleReplaceRule2: &r}
}

func (r *GalicianBarbarismsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
