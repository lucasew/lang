package bitext

import (
	"unicode/utf16"

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
		// Java: last.getStartPos() + last.getToken().length()
		toks := target.GetTokens()
		if len(toks) == 0 {
			return []*rules.RuleMatch{rules.NewRuleMatch(r, target, 0, 0, r.GetMessage())}
		}
		last := toks[len(toks)-1]
		end := last.GetStartPos() + javaStringLen(last.GetToken())
		return []*rules.RuleMatch{rules.NewRuleMatch(r, target, 0, end, r.GetMessage())}
	}
	return nil
}

// javaStringLen matches Java String.length() (UTF-16 code units).
func javaStringLen(s string) int {
	return len(utf16.Encode([]rune(s)))
}

func isLengthDifferent(src, trg string) bool {
	// Java: ((double) src.length() / (double) trg.length()) * 100
	// Empty trg → Infinity (or NaN if both empty); Infinity > MAX_SKEW.
	srcLen := float64(javaStringLen(src))
	trgLen := float64(javaStringLen(trg))
	skew := (srcLen / trgLen) * 100.0
	return skew > maxSkew || skew < minSkew
}

var _ BitextRule = (*DifferentLengthRule)(nil)
