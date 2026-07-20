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
	// Column/EndColumn/Line/EndLine port Tools.checkBitext position adjustments.
	Column    int
	EndColumn int
	Line      int
	EndLine   int
}

// CheckBitext ports Tools.checkBitext for plain strings using AnalyzePlain
// (no POS — inflected false friends fail closed until a tagger is used).
// When rulesList is nil, uses RelevantBitextRules().
func CheckBitext(sourceText, targetText string, rulesList []BitextRule) []BitextMatch {
	src := languagetool.AnalyzePlain(sourceText)
	trg := languagetool.AnalyzePlain(targetText)
	return CheckBitextAnalyzed(src, trg, targetText, rulesList)
}

// CheckBitextAnalyzed ports the bitext-rule half of Tools.checkBitext after
// srcLt.getAnalyzedSentence / trgLt.getAnalyzedSentence.
// targetText is used for Java endColumn default (trg.length()+1 when endColumn < 0).
// Note: Java also runs target monolingual checkAnalyzedSentence first; callers
// that need that path must merge monolingual matches themselves.
func CheckBitextAnalyzed(src, trg *languagetool.AnalyzedSentence, targetText string, rulesList []BitextRule) []BitextMatch {
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
			// Java Tools.checkBitext: adjust positions for bitext rules
			col, endCol, line, endLine := m.Column, m.EndColumn, m.Line, m.EndLine
			// Go RuleMatch zero means unset (Java uses -1 for unset)
			if col < 0 {
				col = 1
			}
			if endCol < 0 {
				endCol = utf16Len(targetText) + 1 // Java counts from 0 for length+1
			}
			if line < 0 {
				line = 1
			}
			if endLine < 0 {
				endLine = 1
			}
			// When Column was never set (0), leave as 0 — Java only adjusts < 0.
			// RuleMatch defaults may be 0 not -1 in Go; preserve FromPos/ToPos as authority.
			out = append(out, BitextMatch{
				RuleID:    id,
				Message:   m.Message,
				FromPos:   m.FromPos,
				ToPos:     m.ToPos,
				Column:    col,
				EndColumn: endCol,
				Line:      line,
				EndLine:   endLine,
			})
		}
	}
	return out
}

func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x10000 {
			n += 2
		} else {
			n++
		}
	}
	return n
}
