package bitext

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

const (
	maxSkew = 250
	minSkew = 30
)

// DifferentLengthRule ports org.languagetool.rules.bitext.DifferentLengthRule.
type DifferentLengthRule struct {
	BitextRuleBase
}

func NewDifferentLengthRule() *DifferentLengthRule {
	return &DifferentLengthRule{BitextRuleBase: BitextRuleBase{
		ID:          "TRANSLATION_LENGTH",
		Description: "Check if translation length is similar to source length",
		Message:     "Source and target translation lengths are very different",
		IssueType:   "length",
	}}
}

func (r *DifferentLengthRule) MatchBitext(source, target *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if source == nil || target == nil {
		return nil
	}
	if isLengthDifferent(source.GetText(), target.GetText()) {
		end := targetEndPos(target)
		return []*rules.RuleMatch{rules.NewRuleMatch(r, target, 0, end, r.GetMessage())}
	}
	return nil
}

func isLengthDifferent(src, trg string) bool {
	if len(trg) == 0 {
		return len(src) > 0
	}
	skew := (float64(len(src)) / float64(len(trg))) * 100.0
	return skew > maxSkew || skew < minSkew
}

var _ BitextRule = (*DifferentLengthRule)(nil)
