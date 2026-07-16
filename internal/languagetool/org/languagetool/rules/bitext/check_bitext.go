package bitext

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// BitextMatch is a simplified match from CheckBitext.
type BitextMatch struct {
	RuleID  string
	Message string
	FromPos int
	ToPos   int
}

// CheckBitext ports Tools.checkBitext for analyzed source/target pairs.
// When rulesList is nil, uses RelevantBitextRules().
func CheckBitext(sourceText, targetText string, rulesList []BitextRule) []BitextMatch {
	src := languagetool.AnalyzePlain(sourceText)
	trg := languagetool.AnalyzePlain(targetText)
	if rulesList == nil {
		rulesList = RelevantBitextRules()
	}
	out := make([]BitextMatch, 0)
	for _, r := range rulesList {
		if r == nil {
			continue
		}
		for _, m := range r.MatchBitext(src, trg) {
			if m == nil {
				continue
			}
			id := r.GetID()
			if rr, ok := m.Rule.(interface{ GetID() string }); ok && rr.GetID() != "" {
				id = rr.GetID()
			}
			out = append(out, BitextMatch{
				RuleID:  id,
				Message: m.Message,
				FromPos: m.FromPos,
				ToPos:   m.ToPos,
			})
		}
	}
	return out
}
