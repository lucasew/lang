package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestFilter_ExclusiveMarkerTargetsFirst(t *testing.T) {
	// Java LET_GO style: <marker>let</marker> + go → FILTER fromPos = "let"
	// (matcher shrinks FromPos/ToPos to InsideMarker tokens).
	let := patterns.NewPatternToken("let", false, false, false)
	let.SetInsideMarker(true)
	goTok := patterns.NewPatternToken("go", false, false, false)
	goTok.SetInsideMarker(false)
	rule := NewDisambiguationPatternRule("LET_GO", "t", "en",
		[]*patterns.PatternToken{let, goTok}, "VB.*", nil, ActionFilter)

	vbTag, nnTag := "VB", "NN"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strp2(languagetool.SentenceStartTagName), nil)),
		func() *languagetool.AnalyzedTokenReadings {
			r := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("let", &vbTag, nil))
			r.AddReading(languagetool.NewAnalyzedToken("let", &nnTag, nil), "test")
			return r
		}(),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("go", &vbTag, nil)),
	}
	pos := 0
	for _, t := range toks {
		t.SetStartPos(pos)
		pos += len(t.GetToken()) + 1
	}
	out := rule.Replace(languagetool.NewAnalyzedSentence(toks))
	var letPOS []string
	for _, tok := range out.GetTokensWithoutWhitespace() {
		if tok.GetToken() != "let" {
			continue
		}
		for _, r := range tok.GetReadings() {
			if r != nil && r.GetPOSTag() != nil {
				letPOS = append(letPOS, *r.GetPOSTag())
			}
		}
	}
	require.Contains(t, letPOS, "VB")
	require.NotContains(t, letPOS, "NN", "FILTER fromPos (marker) on let should drop NN")
}

func TestFilterAll_UsesEachPatternTokenPOS(t *testing.T) {
	// Java TE_X style: te + <marker postag="WKW:.*|ENM:.*"/> → FILTERALL keeps
	// only readings matching that pattern token POS on the marked token.
	te := patterns.NewPatternToken("te", false, false, false)
	te.SetInsideMarker(false)
	verb := patterns.NewPatternToken("", false, false, false)
	verb.SetPosToken(patterns.PosToken{PosTag: "WKW:.*", Regexp: true})
	verb.SetInsideMarker(true)
	rule := NewDisambiguationPatternRule("TE_X", "t", "nl",
		[]*patterns.PatternToken{te, verb}, "", nil, ActionFilterAll)

	wkw, bnw := "WKW:TGW:INF", "BNW:STL:ONV"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strp2(languagetool.SentenceStartTagName), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("te", strp2("VZ"), nil)),
		func() *languagetool.AnalyzedTokenReadings {
			r := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("paard", &wkw, nil))
			r.AddReading(languagetool.NewAnalyzedToken("paard", &bnw, nil), "test")
			return r
		}(),
	}
	pos := 0
	for _, t := range toks {
		t.SetStartPos(pos)
		pos += len(t.GetToken()) + 1
	}
	out := rule.Replace(languagetool.NewAnalyzedSentence(toks))
	var tags []string
	for _, tok := range out.GetTokensWithoutWhitespace() {
		if tok.GetToken() != "paard" {
			continue
		}
		for _, r := range tok.GetReadings() {
			if r != nil && r.GetPOSTag() != nil {
				tags = append(tags, *r.GetPOSTag())
			}
		}
	}
	require.Contains(t, tags, "WKW:TGW:INF")
	require.NotContains(t, tags, "BNW:STL:ONV", "FILTERALL should drop non-matching POS")
}

func TestFilter_MarkerOnSecondToken_JavaFromPos(t *testing.T) {
	// Soft modal style, Java-faithful: will + <marker>run</marker> → FILTER run.
	will := patterns.NewPatternToken("will", false, false, false)
	will.SetInsideMarker(false)
	run := patterns.NewPatternToken("run", false, false, false)
	run.SetInsideMarker(true)
	rule := NewDisambiguationPatternRule("SOFT_WILL_RUN_VB", "t", "en",
		[]*patterns.PatternToken{will, run}, "VB", nil, ActionFilter)

	vbTag, nnTag, mdTag := "VB", "NN", "MD"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strp2(languagetool.SentenceStartTagName), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("will", &mdTag, nil)),
		func() *languagetool.AnalyzedTokenReadings {
			r := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("run", &vbTag, nil))
			r.AddReading(languagetool.NewAnalyzedToken("run", &nnTag, nil), "test")
			return r
		}(),
	}
	pos := 0
	for _, t := range toks {
		t.SetStartPos(pos)
		pos += len(t.GetToken()) + 1
	}
	out := rule.Replace(languagetool.NewAnalyzedSentence(toks))
	var runPOS []string
	for _, tok := range out.GetTokensWithoutWhitespace() {
		if tok.GetToken() != "run" {
			continue
		}
		for _, r := range tok.GetReadings() {
			if r != nil && r.GetPOSTag() != nil {
				runPOS = append(runPOS, *r.GetPOSTag())
			}
		}
	}
	require.Contains(t, runPOS, "VB")
	require.NotContains(t, runPOS, "NN", "FILTER fromPos on marked run should drop NN")
}

func strp2(s string) *string { return &s }
