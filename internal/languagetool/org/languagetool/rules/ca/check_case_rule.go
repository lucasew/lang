package ca

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
			ID:                   "CA_CHECKCASE",
			Description:          "Comprova majúscules i minúscules: $match",
			ShortMsg:             "Majúscules i minúscules",
			MessageTemplate:      "Majúscules i minúscules recomanades. Alguns llibres d'estil poden suggerir solucions diferents en alguns casos.",
			SuggestionsSeparator: " o ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "ca",
			SubRuleSpecificIDs:   true,
			CheckingCase:         true,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/ca/check_case.txt"); err != nil {
			panic(err)
		}
		checkCaseBase = base
	})
	return checkCaseBase
}

// CheckCaseRule ports org.languagetool.rules.ca.CheckCaseRule
// (AbstractCheckCaseRule → ASR2 with CheckingCase).
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
