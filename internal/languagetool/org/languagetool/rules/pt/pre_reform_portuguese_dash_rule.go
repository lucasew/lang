package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/pre-reform-compounds.txt
var preReformCompoundsFS embed.FS

var (
	preDashOnce     sync.Once
	preDashPatterns []string
)

func loadPreReformDashPatterns() []string {
	preDashOnce.Do(func() {
		f, err := preReformCompoundsFS.Open("data/pre-reform-compounds.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		p, err := rules.LoadDashCompoundPatterns(f)
		if err != nil {
			panic(err)
		}
		preDashPatterns = p
	})
	return preDashPatterns
}

// PreReformPortugueseDashRule ports org.languagetool.rules.pt.PreReformPortugueseDashRule.
type PreReformPortugueseDashRule struct {
	*rules.AbstractDashRule
}

func NewPreReformPortugueseDashRule(messages map[string]string) *PreReformPortugueseDashRule {
	base := &rules.AbstractDashRule{
		Messages:         messages,
		ID:               "PT_PREAO_DASH_RULE",
		CompoundPatterns: loadPreReformDashPatterns(),
		Message:          "Um travessão foi utilizado em vez de um hífen.",
		Description:      "Travessões no lugar de hífens (pré-Acordo Ortográfico)",
		IsLetter:         isPortugueseLetter,
	}
	return &PreReformPortugueseDashRule{AbstractDashRule: base}
}

func (r *PreReformPortugueseDashRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractDashRule.Match(sentence)
}
