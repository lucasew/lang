package bitext

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SameTranslationRule ports org.languagetool.rules.bitext.SameTranslationRule.
type SameTranslationRule struct {
	BitextRuleBase
}

func NewSameTranslationRule() *SameTranslationRule {
	return &SameTranslationRule{BitextRuleBase: BitextRuleBase{
		ID:          "SAME_TRANSLATION",
		Description: "Check if translation is the same as source",
		Message:     "Source and target translation are the same",
		IssueType:   "untranslated",
	}}
}

func (r *SameTranslationRule) MatchBitext(source, target *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if source == nil || target == nil {
		return nil
	}
	// Java: tokens without whitespace length > 3 and texts equal
	if len(source.GetTokensWithoutWhitespace()) > 3 && source.GetText() == target.GetText() {
		end := targetEndPos(target)
		return []*rules.RuleMatch{rules.NewRuleMatch(r, target, 1, end, r.GetMessage())}
	}
	return nil
}

var _ BitextRule = (*SameTranslationRule)(nil)
