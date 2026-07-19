package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CatalanRemoteRewriteFilter ports org.languagetool.rules.ca.CatalanRemoteRewriteFilter.
// Remote HTTP is pluggable via Rewrite (Java: sendPostRequest).
type CatalanRemoteRewriteFilter struct {
	// Rewrite returns a corrected sentence for ruleID, or "" on failure.
	Rewrite func(sentence, ruleID string) string
}

func NewCatalanRemoteRewriteFilter() *CatalanRemoteRewriteFilter {
	return &CatalanRemoteRewriteFilter{}
}

// AcceptRuleMatch ports CatalanRemoteRewriteFilter.acceptRuleMatch.
// Args: optional suppressMatch=true to drop match when rewrite fails.
func (f *CatalanRemoteRewriteFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	suppress := strings.EqualFold(arguments["suppressMatch"], "true")
	if match.Sentence == nil {
		if suppress {
			return nil
		}
		return match
	}
	orig := strings.TrimSpace(match.Sentence.GetText())
	ruleID := ruleIDFromMatch(match)
	res := f.Apply(orig, match.GetFromPos(), match.GetToPos(), ruleID, suppress)
	if !res.Keep {
		return nil
	}
	if len(res.Replacements) == 0 {
		// Rewrite unavailable: keep original match unless suppress.
		return match
	}
	if res.ToPos <= res.FromPos {
		// Java throws IllegalArgumentException; fail-closed drop.
		return nil
	}
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, res.FromPos, res.ToPos, match.GetMessage())
	out.ShortMessage = match.ShortMessage
	out.SetSuggestedReplacements(res.Replacements)
	return out
}

// RemoteRewriteResult holds the expanded match range and suggestions.
type RemoteRewriteResult struct {
	FromPos, ToPos int
	Replacements   []string
	Keep           bool
}

// Apply maps a remote correction onto the original match span using DiffsAsMatches.
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

func ruleIDFromMatch(match *rules.RuleMatch) string {
	if match == nil || match.GetRule() == nil {
		return ""
	}
	if r, ok := match.GetRule().(interface{ GetID() string }); ok {
		return r.GetID()
	}
	return ""
}
