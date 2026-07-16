package ga

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/specific_case.txt
var specificCaseFS embed.FS

var (
	specificCaseOnce sync.Once
	specificCaseMap  map[string]string
	specificCaseMax  int
)

func loadSpecificCase() (map[string]string, int) {
	specificCaseOnce.Do(func() {
		f, err := specificCaseFS.Open("data/specific_case.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, maxLen, err := rules.LoadSpecificCasePhrases(f)
		if err != nil {
			panic(err)
		}
		specificCaseMap = m
		specificCaseMax = maxLen
	})
	return specificCaseMap, specificCaseMax
}

// IrishSpecificCaseRule ports org.languagetool.rules.ga.IrishSpecificCaseRule.
type IrishSpecificCaseRule struct {
	*rules.AbstractSpecificCaseRule
}

func NewIrishSpecificCaseRule(messages map[string]string) *IrishSpecificCaseRule {
	m, maxLen := loadSpecificCase()
	base := &rules.AbstractSpecificCaseRule{
		Messages:                   messages,
		LcToProper:                 m,
		MaxPhraseLen:               maxLen,
		ID:                         "GA_SPECIFIC_CASE",
		Description:                "Checks upper/lower case spelling of some proper nouns",
		InitialCapitalMessage:      "Más ainmfhocal dílis é, scríobh é i gceannlitreacha.",
		OtherCapitalizationMessage: "If the term is a proper noun, use the suggested capitalization.",
		ShortMsg:                   "Special capitalization",
	}
	return &IrishSpecificCaseRule{AbstractSpecificCaseRule: base}
}

func (r *IrishSpecificCaseRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSpecificCaseRule.Match(sentence)
}
