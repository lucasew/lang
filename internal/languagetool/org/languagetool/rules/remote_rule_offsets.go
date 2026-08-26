package rules

import (
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// ComputeOffsetShifts ports RemoteRule.computeOffsetShifts.
// Maps code-point indices → Java UTF-16 string indices (for emoji/surrogate pairs).
func ComputeOffsetShifts(s string) []int {
	u16 := utf16.Encode([]rune(s))
	lenU16 := len(u16)
	offsets := make([]int, lenU16+1)
	shifted := 0 // UTF-16 index
	original := 0 // code-point index
	for _, r := range s {
		if original < len(offsets) {
			offsets[original] = shifted
		}
		if r > 0xFFFF {
			shifted += 2
		} else {
			shifted++
		}
		original++
	}
	if original < len(offsets) {
		offsets[original] = shifted
	}
	for i := original + 1; i < len(offsets); i++ {
		offsets[i] = offsets[i-1] + 1
	}
	return offsets
}

// FixMatchOffsets ports RemoteRule.fixMatchOffsets — rewrites match positions
// from code-point offsets to UTF-16 offsets based on sentence text.
func FixMatchOffsets(sentence *languagetool.AnalyzedSentence, matches []*RuleMatch) {
	if sentence == nil || len(matches) == 0 {
		return
	}
	text := sentence.GetText()
	shifts := ComputeOffsetShifts(text)
	for _, m := range matches {
		if m == nil {
			continue
		}
		from, to := m.FromPos, m.ToPos
		if from >= 0 && from < len(shifts) {
			from = shifts[from]
		}
		if to >= 0 && to < len(shifts) {
			to = shifts[to]
		}
		m.SetOffsetPosition(from, to)
	}
}
