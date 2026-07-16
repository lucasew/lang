package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CatalanRemoteRewriteFilter ports post-processing from
// org.languagetool.rules.ca.CatalanRemoteRewriteFilter.
// Remote HTTP is pluggable via Rewrite.
type CatalanRemoteRewriteFilter struct {
	// Rewrite returns a corrected sentence for ruleID, or "" on failure.
	Rewrite func(sentence, ruleID string) string
}

func NewCatalanRemoteRewriteFilter() *CatalanRemoteRewriteFilter {
	return &CatalanRemoteRewriteFilter{}
}

// RemoteRewriteResult holds the expanded match range and suggestions.
type RemoteRewriteResult struct {
	FromPos, ToPos int
	Replacements   []string
	Keep           bool
}

// Apply maps a remote correction onto the original match span using DiffsAsMatches.
// fromPos/toPos are the pattern match offsets into originalSentence.
// suppressMatch controls behaviour when rewrite is empty or unusable.
func (f *CatalanRemoteRewriteFilter) Apply(originalSentence string, fromPos, toPos int, ruleID string, suppressMatch bool) RemoteRewriteResult {
	trim := func(s string) string { return strings.TrimSpace(s) }
	orig := trim(originalSentence)
	if f.Rewrite == nil {
		return RemoteRewriteResult{Keep: !suppressMatch, FromPos: fromPos, ToPos: toPos}
	}
	corrected := f.Rewrite(orig, ruleID)
	if corrected == "" {
		return RemoteRewriteResult{Keep: !suppressMatch, FromPos: fromPos, ToPos: toPos}
	}
	diffs := tools.NewDiffsAsMatches()
	pseudo := diffs.GetPseudoMatches(orig, corrected)
	joined := diffs.GetJoinedMatch(pseudo, orig, fromPos-2, toPos+60)
	if joined == nil {
		return RemoteRewriteResult{Keep: !suppressMatch, FromPos: fromPos, ToPos: toPos}
	}
	reps := joined.GetReplacements()
	if len(reps) == 0 {
		return RemoteRewriteResult{Keep: !suppressMatch, FromPos: fromPos, ToPos: toPos}
	}
	suggestion := reps[0]
	jFrom, jTo := joined.GetFromPos(), joined.GetToPos()
	if jTo > len(orig) || jFrom < 0 || jTo < jFrom {
		return RemoteRewriteResult{Keep: !suppressMatch, FromPos: fromPos, ToPos: toPos}
	}
	underlined := orig[jFrom:jTo]
	if (jTo == len(orig) || jFrom == 0) && trim(underlined) == "" {
		return RemoteRewriteResult{Keep: !suppressMatch, FromPos: fromPos, ToPos: toPos}
	}
	if trim(suggestion) == trim(underlined) {
		return RemoteRewriteResult{Keep: !suppressMatch, FromPos: fromPos, ToPos: toPos}
	}
	return RemoteRewriteResult{
		FromPos:      jFrom,
		ToPos:        jTo,
		Replacements: reps,
		Keep:         true,
	}
}
