package fr

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PostponedAdjectiveConcordanceFilter ports
// org.languagetool.rules.fr.PostponedAdjectiveConcordanceFilter (simplified).
// French FreeLing-style POS tags are space-separated: "J m s", "N f p", etc.
type PostponedAdjectiveConcordanceFilter struct {
	MaxLevels int
}

func NewPostponedAdjectiveConcordanceFilter() *PostponedAdjectiveConcordanceFilter {
	return &PostponedAdjectiveConcordanceFilter{MaxLevels: 4}
}

var (
	frAdjMS = regexp.MustCompile(`^J [me] (s|sp)$|^V ppa m s$`)
	frAdjFS = regexp.MustCompile(`^J [fe] (s|sp)$|^V ppa f s$`)
	frAdjMP = regexp.MustCompile(`^J [me] (p|sp)$|^V ppa m p$`)
	frAdjFP = regexp.MustCompile(`^J [fe] (p|sp)$|^V ppa f p$`)
	frNomMS = regexp.MustCompile(`^[NZ] m s$`)
	frNomFS = regexp.MustCompile(`^[NZ] f s$`)
	frNomMP = regexp.MustCompile(`^[NZ] m p$`)
	frNomFP = regexp.MustCompile(`^[NZ] f p$`)
	frDetMS = regexp.MustCompile(`^(P\+)?D m s$`)
	frDetFS = regexp.MustCompile(`^(P\+)?D f s$`)
	frDetMP = regexp.MustCompile(`^(P\+)?D m p$`)
	frDetFP = regexp.MustCompile(`^(P\+)?D f p$`)
)

type frGN struct{ m, f, p, s bool }

func frGNFrom(tag string) frGN {
	var g frGN
	switch {
	case frAdjMS.MatchString(tag) || frNomMS.MatchString(tag) || frDetMS.MatchString(tag):
		g.m, g.s = true, true
	case frAdjFS.MatchString(tag) || frNomFS.MatchString(tag) || frDetFS.MatchString(tag):
		g.f, g.s = true, true
	case frAdjMP.MatchString(tag) || frNomMP.MatchString(tag) || frDetMP.MatchString(tag):
		g.m, g.p = true, true
	case frAdjFP.MatchString(tag) || frNomFP.MatchString(tag) || frDetFP.MatchString(tag):
		g.f, g.p = true, true
	}
	return g
}

func (g frGN) agrees(o frGN) bool {
	if (g.m && o.m) || (g.f && o.f) {
		if (g.s && o.s) || (g.p && o.p) {
			return true
		}
	}
	if !g.m && !g.f {
		return true
	}
	if !o.m && !o.f {
		return true
	}
	return false
}

// AcceptRuleMatch keeps the match when the adjective disagrees with preceding
// noun/det gender-number in a short lookback window.
func (f *PostponedAdjectiveConcordanceFilter) AcceptRuleMatch(
	match *rules.RuleMatch,
	_ map[string]string,
	patternTokens []*languagetool.AnalyzedTokenReadings,
) *rules.RuleMatch {
	if match == nil || len(patternTokens) == 0 {
		return nil
	}
	adjIdx := len(patternTokens) - 1
	adjGN := frCollect(patternTokens[adjIdx])
	if !adjGN.m && !adjGN.f {
		return match
	}
	levels := f.MaxLevels
	if levels <= 0 {
		levels = 4
	}
	start := adjIdx - levels*2
	if start < 0 {
		start = 0
	}
	for i := start; i < adjIdx; i++ {
		g := frCollect(patternTokens[i])
		if g.agrees(adjGN) && (g.m || g.f) {
			return nil
		}
	}
	return match
}

func frCollect(tok *languagetool.AnalyzedTokenReadings) frGN {
	var out frGN
	if tok == nil {
		return out
	}
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		g := frGNFrom(*r.GetPOSTag())
		out.m = out.m || g.m
		out.f = out.f || g.f
		out.s = out.s || g.s
		out.p = out.p || g.p
	}
	return out
}
