package en

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

// EnglishSpecificCaseRule ports org.languagetool.rules.en.EnglishSpecificCaseRule.
type EnglishSpecificCaseRule struct {
	*rules.AbstractSpecificCaseRule
}

func NewEnglishSpecificCaseRule(messages map[string]string) *EnglishSpecificCaseRule {
	m, maxLen := loadSpecificCase()
	base := &rules.AbstractSpecificCaseRule{
		Messages:                   messages,
		LcToProper:                 m,
		MaxPhraseLen:               maxLen,
		ID:                         "EN_SPECIFIC_CASE",
		Description:                "Checks upper/lower case spelling of some proper nouns",
		InitialCapitalMessage:      "If the term is a proper noun, use initial capitals.",
		OtherCapitalizationMessage: "If the term is a proper noun, use the suggested capitalization.",
		ShortMsg:                   "Proper noun",
		// Java EnglishSpecificCaseRule.setUrl capital-letters insights
		URL: "https://languagetool.org/insights/post/spelling-capital-letters/",
	}
	rules.InitSpecificCaseMeta(base, messages)
	// Java: addExamplePair(Harry potter → Harry Potter)
	base.AddExamplePair(
		rules.Wrong("I really like <marker>Harry potter</marker>."),
		rules.Fixed("I really like <marker>Harry Potter</marker>."),
	)
	return &EnglishSpecificCaseRule{AbstractSpecificCaseRule: base}
}

func (r *EnglishSpecificCaseRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSpecificCaseRule.Match(sentence)
}
