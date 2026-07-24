package km

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

var kmConjunctions = regexp.MustCompile(`^(ដើម្បី|និង|ពីព្រោះ)$`)

// KhmerSpaceBeforeRule ports org.languagetool.rules.km.KhmerSpaceBeforeRule.
type KhmerSpaceBeforeRule struct {
	*rules.AbstractSpaceBeforeRule
}

func NewKhmerSpaceBeforeRule(messages map[string]string) *KhmerSpaceBeforeRule {
	base := &rules.AbstractSpaceBeforeRule{
		ID:           "KM_SPACE_BEFORE_CONJUNCTION",
		Description:  "Checks for missing space before some conjunctions",
		ShortMsg:     "Missing white space",
		Suggestion:   "Missing white space before conjunction",
		Conjunctions: kmConjunctions,
	}
	rules.InitSpaceBeforeMeta(base, messages)
	return &KhmerSpaceBeforeRule{AbstractSpaceBeforeRule: base}
}

func (r *KhmerSpaceBeforeRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSpaceBeforeRule.Match(sentence)
}
