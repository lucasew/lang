package be

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

// BelarusianSpecificCaseRule ports org.languagetool.rules.be.BelarusianSpecificCaseRule.
type BelarusianSpecificCaseRule struct {
	*rules.AbstractSpecificCaseRule
}

func NewBelarusianSpecificCaseRule(messages map[string]string) *BelarusianSpecificCaseRule {
	m, maxLen := loadSpecificCase()
	base := &rules.AbstractSpecificCaseRule{
		Messages:                   messages,
		LcToProper:                 m,
		MaxPhraseLen:               maxLen,
		ID:                         "BE_SPECIFIC_CASE",
		Description:                "Напісанне спецыяльных найменняў у верхнім або ніжнім рэгістры",
		InitialCapitalMessage:      "Уласныя імёны і назвы пішуцца з вялікай літары.",
		OtherCapitalizationMessage: "Калі гэта уласнае імя або назва, выкарыстоўвайце прапанаванае напісанне.",
		ShortMsg:                   "Proper noun",
	}
	return &BelarusianSpecificCaseRule{AbstractSpecificCaseRule: base}
}

func (r *BelarusianSpecificCaseRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSpecificCaseRule.Match(sentence)
}
