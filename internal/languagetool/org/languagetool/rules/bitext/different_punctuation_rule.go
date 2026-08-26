package bitext

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// DifferentPunctuationRule ports org.languagetool.rules.bitext.DifferentPunctuationRule.
type DifferentPunctuationRule struct {
	BitextRuleBase
}

func NewDifferentPunctuationRule() *DifferentPunctuationRule {
	return &DifferentPunctuationRule{BitextRuleBase: BitextRuleBase{
		ID:          "DIFFERENT_PUNCTUATION",
		Description: "Check if translation has ending punctuation different from the source",
		Message:     "Source and target translation have different ending punctuation",
		IssueType:   "typographical",
	}}
}

func (r *DifferentPunctuationRule) MatchBitext(source, target *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if source == nil || target == nil {
		return nil
	}
	st := source.GetTokens()
	tt := target.GetTokens()
	if len(st) == 0 || len(tt) == 0 {
		return nil
	}
	lastT := tt[len(tt)-1]
	lastS := st[len(st)-1]
	lastTok := lastT.GetToken()
	if (lastTok == "." || lastTok == "?" || lastTok == "!") && lastTok != lastS.GetToken() {
		return []*rules.RuleMatch{rules.NewRuleMatch(r, target, 1, lastT.GetEndPos(), r.GetMessage())}
	}
	return nil
}

var _ BitextRule = (*DifferentPunctuationRule)(nil)
