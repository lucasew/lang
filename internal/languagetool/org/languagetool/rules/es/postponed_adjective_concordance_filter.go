package es

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PostponedAdjectiveConcordanceFilter ports Spanish postponed adj concordance
// (same POS family as Catalan FreeLing tags).
type PostponedAdjectiveConcordanceFilter struct {
	MaxLevels int
}

func NewPostponedAdjectiveConcordanceFilter() *PostponedAdjectiveConcordanceFilter {
	return &PostponedAdjectiveConcordanceFilter{MaxLevels: 4}
}

var (
	esAdjMS = regexp.MustCompile(`^A..[MC][SN]|^V\.P\.\.SM`)
	esAdjFS = regexp.MustCompile(`^A..[FC][SN]|^V\.P\.\.SF`)
	esAdjMP = regexp.MustCompile(`^A..[MC][PN]|^V\.P\.\.PM`)
	esAdjFP = regexp.MustCompile(`^A..[FC][PN]|^V\.P\.\.PF`)
	esNomMS = regexp.MustCompile(`^N.MS`)
	esNomFS = regexp.MustCompile(`^N.FS`)
	esNomMP = regexp.MustCompile(`^N.MP`)
	esNomFP = regexp.MustCompile(`^N.FP`)
)

type esGN struct{ m, f, p, s bool }

func esGNFrom(tag string) esGN {
	var g esGN
	switch {
	case esAdjMS.MatchString(tag) || esNomMS.MatchString(tag):
		g.m, g.s = true, true
	case esAdjFS.MatchString(tag) || esNomFS.MatchString(tag):
		g.f, g.s = true, true
	case esAdjMP.MatchString(tag) || esNomMP.MatchString(tag):
		g.m, g.p = true, true
	case esAdjFP.MatchString(tag) || esNomFP.MatchString(tag):
		g.f, g.p = true, true
	}
	return g
}

func (g esGN) agrees(o esGN) bool {
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

func (f *PostponedAdjectiveConcordanceFilter) AcceptRuleMatch(
	match *rules.RuleMatch,
	_ map[string]string,
	patternTokens []*languagetool.AnalyzedTokenReadings,
) *rules.RuleMatch {
	if match == nil || len(patternTokens) == 0 {
		return nil
	}
	adjIdx := len(patternTokens) - 1
	adjGN := esCollect(patternTokens[adjIdx])
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
		if esCollect(patternTokens[i]).agrees(adjGN) && (esCollect(patternTokens[i]).m || esCollect(patternTokens[i]).f) {
			return nil
		}
	}
	return match
}

func esCollect(tok *languagetool.AnalyzedTokenReadings) esGN {
	var out esGN
	if tok == nil {
		return out
	}
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		g := esGNFrom(*r.GetPOSTag())
		out.m = out.m || g.m
		out.f = out.f || g.f
		out.s = out.s || g.s
		out.p = out.p || g.p
	}
	return out
}
