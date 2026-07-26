package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/pre-reform-compounds.txt
var preReformCompoundsFS embed.FS

// preReformCompoundsResourcePath is the official Java resource path loaded by
// PreReformPortugueseDashRule.getCompoundsData (JLanguageTool resource dir).
const preReformCompoundsResourcePath = "/pt/pre-reform-compounds.txt"

var (
	preDashOnce     sync.Once
	preDashPatterns []string
)

func loadPreReformDashPatterns() []string {
	preDashOnce.Do(func() {
		// Java: loadCompoundFile("/pt/pre-reform-compounds.txt") once (DCL / sync.Once).
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
	// Java: super(messages); setLocQualityIssueType(ITSIssueType.Typographical);
	// No setUrl / addExamplePair in Java PT pre-reform dash rule.
	// getDescription is exactly "Travessões no lugar de hífens" (no pré-Acordo suffix).
	base := &rules.AbstractDashRule{
		ID:               "PT_PREAO_DASH_RULE",
		CompoundPatterns: loadPreReformDashPatterns(),
		Message:          "Um travessão foi utilizado em vez de um hífen.",
		Description:      "Travessões no lugar de hífens",
		IsLetter:         isPortugueseLetter, // shared with PostReform (identical Java PATTERN)
		IssueType:        rules.ITSTypographical,
	}
	rules.InitDashRuleMeta(base, messages)
	return &PreReformPortugueseDashRule{AbstractDashRule: base}
}

func (r *PreReformPortugueseDashRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractDashRule.Match(sentence)
}
