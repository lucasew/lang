package en

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/compounds.txt
var compoundsFS embed.FS

var (
	dashOnce     sync.Once
	dashPatterns []string
)

func loadDashPatterns() []string {
	dashOnce.Do(func() {
		f, err := compoundsFS.Open("data/compounds.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		p, err := rules.LoadDashCompoundPatterns(f)
		if err != nil {
			panic(err)
		}
		dashPatterns = p
	})
	return dashPatterns
}

// EnglishDashRule ports org.languagetool.rules.en.EnglishDashRule.
type EnglishDashRule struct {
	*rules.AbstractDashRule
}

func NewEnglishDashRule(messages map[string]string) *EnglishDashRule {
	base := &rules.AbstractDashRule{
		ID:               "EN_DASH_RULE",
		CompoundPatterns: loadDashPatterns(),
		Message:          "A dash was used instead of a hyphen.",
		// Java EnglishDashRule.getDescription
		Description: "Checks if hyphenated words were spelled with dashes (e.g., 'T — shirt' instead 'T-shirt').",
		// Java setUrl hyphen insights
		URL: "https://languagetool.org/insights/post/hyphen/",
	}
	rules.InitDashRuleMeta(base, messages)
	// Java: addExamplePair(T—shirt → T-shirt)
	base.AddExamplePair(
		rules.Wrong("I'll buy a new <marker>T—shirt</marker>."),
		rules.Fixed("I'll buy a new <marker>T-shirt</marker>."),
	)
	return &EnglishDashRule{AbstractDashRule: base}
}

func (r *EnglishDashRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractDashRule.Match(sentence)
}
