package fa

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

var faConjunctions = regexp.MustCompile(`^(و|به|با|تا|زیرا|چون|بنابراین|چونکه)$`)

// PersianSpaceBeforeRule ports org.languagetool.rules.fa.PersianSpaceBeforeRule.
type PersianSpaceBeforeRule struct {
	*rules.AbstractSpaceBeforeRule
}

func NewPersianSpaceBeforeRule(messages map[string]string) *PersianSpaceBeforeRule {
	base := &rules.AbstractSpaceBeforeRule{
		Messages:     messages,
		ID:           "FA_SPACE_BEFORE_CONJUNCTION",
		Description:  "Checks for missing space before some conjunctions",
		ShortMsg:     "Missing white space",
		Suggestion:   "Missing white space before conjunction",
		Conjunctions: faConjunctions,
	}
	return &PersianSpaceBeforeRule{AbstractSpaceBeforeRule: base}
}

func (r *PersianSpaceBeforeRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSpaceBeforeRule.Match(sentence)
}
