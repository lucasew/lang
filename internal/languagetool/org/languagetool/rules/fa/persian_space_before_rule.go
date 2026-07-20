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
		ID: "FA_SPACE_BEFORE_CONJUNCTION",
		// Java PersianSpaceBeforeRule getDescription / getShort / getSuggestion
		Description:  "بررسی‌کردن فاصله قبل از حرف ربط",
		ShortMsg:     "فاصلهٔ حذف‌شده",
		Suggestion:   "فاصلهٔ قبل از حرف ربط حذف شده‌است",
		Conjunctions: faConjunctions,
		// Java setDefaultOff()
		DefaultOff: true,
	}
	rules.InitSpaceBeforeMeta(base, messages)
	return &PersianSpaceBeforeRule{AbstractSpaceBeforeRule: base}
}

func (r *PersianSpaceBeforeRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSpaceBeforeRule.Match(sentence)
}
