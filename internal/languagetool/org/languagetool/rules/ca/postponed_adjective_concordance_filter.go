package ca

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PostponedAdjectiveConcordanceFilter ports a simplified surface of
// org.languagetool.rules.ca.PostponedAdjectiveConcordanceFilter:
// keep the match when the adjective POS does not agree in gender/number
// with any preceding noun/det within a short window.
type PostponedAdjectiveConcordanceFilter struct {
	MaxLevels int
}

func NewPostponedAdjectiveConcordanceFilter() *PostponedAdjectiveConcordanceFilter {
	return &PostponedAdjectiveConcordanceFilter{MaxLevels: 4}
}

var (
	caAdjMS = regexp.MustCompile(`^A..[MC][SN]|^V\.P\.\.SM|^PX\.MS`)
	caAdjFS = regexp.MustCompile(`^A..[FC][SN]|^V\.P\.\.SF|^PX\.FS`)
	caAdjMP = regexp.MustCompile(`^A..[MC][PN]|^V\.P\.\.PM|^PX\.MP`)
	caAdjFP = regexp.MustCompile(`^A..[FC][PN]|^V\.P\.\.PF|^PX\.FP`)
	caNomMS = regexp.MustCompile(`^N.MS|^PI0MS`)
	caNomFS = regexp.MustCompile(`^N.FS|^PI0FS`)
	caNomMP = regexp.MustCompile(`^N.MP`)
	caNomFP = regexp.MustCompile(`^N.FP`)
	caDetMS = regexp.MustCompile(`^D[NDA0IP]0MS0`)
	caDetFS = regexp.MustCompile(`^D[NDA0IP]0FS0`)
	caDetMP = regexp.MustCompile(`^D[NDA0IP]0MP0`)
	caDetFP = regexp.MustCompile(`^D[NDA0IP]0FP0`)
)

type gn struct{ m, f, p, s bool } // gender/number slots

func gnFromTag(tag string) gn {
	var g gn
	switch {
	case caAdjMS.MatchString(tag) || caNomMS.MatchString(tag) || caDetMS.MatchString(tag):
		g.m, g.s = true, true
	case caAdjFS.MatchString(tag) || caNomFS.MatchString(tag) || caDetFS.MatchString(tag):
		g.f, g.s = true, true
	case caAdjMP.MatchString(tag) || caNomMP.MatchString(tag) || caDetMP.MatchString(tag):
		g.m, g.p = true, true
	case caAdjFP.MatchString(tag) || caNomFP.MatchString(tag) || caDetFP.MatchString(tag):
		g.f, g.p = true, true
	}
	return g
}

func (g gn) agrees(o gn) bool {
	// agree if share gender and number where both specify
	if (g.m && o.m) || (g.f && o.f) {
		if (g.s && o.s) || (g.p && o.p) {
			return true
		}
	}
	// empty slots = unknown → treat as agreeing (fail open)
	if !g.m && !g.f && !g.s && !g.p {
		return true
	}
	if !o.m && !o.f && !o.s && !o.p {
		return true
	}
	return false
}

// AcceptRuleMatch keeps match if adjective disagrees with all preceding nouns/dets.
// args may contain "adj" index (1-based in pattern tokens).
func (f *PostponedAdjectiveConcordanceFilter) AcceptRuleMatch(
	match *rules.RuleMatch,
	args map[string]string,
	patternTokens []*languagetool.AnalyzedTokenReadings,
) *rules.RuleMatch {
	if match == nil || len(patternTokens) == 0 {
		return nil
	}
	adjIdx := len(patternTokens) - 1
	if v, ok := args["adj_pos"]; ok {
		// 1-based
		var n int
		for _, ch := range v {
			if ch >= '0' && ch <= '9' {
				n = n*10 + int(ch-'0')
			}
		}
		if n > 0 && n <= len(patternTokens) {
			adjIdx = n - 1
		}
	}
	adj := patternTokens[adjIdx]
	adjGN := collectGN(adj)
	if !adjGN.m && !adjGN.f {
		return match // no adj reading → keep original match decision
	}
	// look back
	levels := f.MaxLevels
	if levels <= 0 {
		levels = 4
	}
	start := adjIdx - levels*2
	if start < 0 {
		start = 0
	}
	for i := start; i < adjIdx; i++ {
		g := collectGN(patternTokens[i])
		if g.agrees(adjGN) && (g.m || g.f) {
			return nil // agrees with something before → discard
		}
	}
	return match
}

func collectGN(tok *languagetool.AnalyzedTokenReadings) gn {
	var out gn
	if tok == nil {
		return out
	}
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		g := gnFromTag(*r.GetPOSTag())
		out.m = out.m || g.m
		out.f = out.f || g.f
		out.s = out.s || g.s
		out.p = out.p || g.p
	}
	return out
}
