package gl

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
			ID:                   "GL_WORDINESS_REPLACE",
			Description:          "2. Expresións prolixas",
			ShortMsg:             "Expresión prolixa",
			MessageTemplate:      "'$match' é unha expresión innecesariamente complexa. É preferíbel dicir $suggestions",
			SuggestionsSeparator: " ou ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "gl",
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/gl/wordiness.txt"); err != nil {
			panic(err)
		}
		wordinessBase = base
	})
	return wordinessBase
}

// GalicianWordinessRule ports org.languagetool.rules.gl.GalicianWordinessRule.
type GalicianWordinessRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewGalicianWordinessRule(messages map[string]string) *GalicianWordinessRule {
	base := loadWordiness()
	r := *base
	r.Messages = messages
	return &GalicianWordinessRule{AbstractSimpleReplaceRule2: &r}
}

func (r *GalicianWordinessRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
