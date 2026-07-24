package nl

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/check_case.txt
var checkCaseFS embed.FS

var (
	checkCaseOnce sync.Once
	checkCaseBase *rules.AbstractSimpleReplaceRule2
)

func loadCheckCase() *rules.AbstractSimpleReplaceRule2 {
	checkCaseOnce.Do(func() {
		f, err := checkCaseFS.Open("data/check_case.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                            "NL_CHECKCASE",
			Description:                   "Controle op hoofd- en kleine letters: $match",
			ShortMsg:                      "Schrijfwijze",
			MessageTemplate:               "Juiste schrijfwijze",
			SuggestionsSeparator:          " of ",
			CaseSens:                      rules.CaseInsensitive,
			LanguageCode:                  "nl",
			SubRuleSpecificIDs:            true,
			CheckingCase:                  true,
			MatchShortAllUpperInCheckCase: true, // setIgnoreShortUppercaseWords(false)
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/nl/check_case.txt"); err != nil {
			panic(err)
		}
		checkCaseBase = base
	})
	return checkCaseBase
}

// CheckCaseRule ports org.languagetool.rules.nl.CheckCaseRule.
type CheckCaseRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewCheckCaseRule(messages map[string]string) *CheckCaseRule {
	base := loadCheckCase()
	r := *base
	r.Messages = messages
	return &CheckCaseRule{AbstractSimpleReplaceRule2: &r}
}

func (r *CheckCaseRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
