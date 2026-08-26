package bitext

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// ToLocalMatches projects bitext matches onto the target side as LocalMatch
// for use with CorrectTextFromLocalMatches / reporting helpers.
func ToLocalMatches(ms []BitextMatch) []languagetool.LocalMatch {
	out := make([]languagetool.LocalMatch, 0, len(ms))
	for _, m := range ms {
		out = append(out, languagetool.LocalMatch{
			FromPos: m.FromPos,
			ToPos:   m.ToPos,
			Message: m.Message,
			RuleID:  m.RuleID,
		})
	}
	return out
}

// CheckBitextLocal runs default bitext rules and returns LocalMatch list.
func CheckBitextLocal(sourceText, targetText string) []languagetool.LocalMatch {
	return ToLocalMatches(CheckBitext(sourceText, targetText, nil))
}
