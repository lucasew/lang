package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func sentStart() *languagetool.AnalyzedTokenReadings {
	tag := languagetool.SentenceStartTagName
	return languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &tag, nil), 0)
}

func TestDisambiguationPatternRuleImmunize(t *testing.T) {
	posNN := "NN"
	toks := []*languagetool.AnalyzedTokenReadings{
		sentStart(),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("foo", &posNN, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("bar", &posNN, nil), 4),
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	rule := NewDisambiguationPatternRule(
		"D1", "immunize foo bar", "en",
		[]*patterns.PatternToken{patterns.Token("foo"), patterns.Token("bar")},
		"", nil, ActionImmunize,
	)
	out := rule.Replace(sent)
	nws := out.GetTokensWithoutWhitespace()
	// find foo
	foundImm := false
	for _, t := range nws {
		if t.GetToken() == "foo" && t.IsImmunized() {
			foundImm = true
		}
	}
	require.True(t, foundImm, "expected foo to be immunized")
}

func TestXmlRuleDisambiguator(t *testing.T) {
	pos := "VB"
	toks := []*languagetool.AnalyzedTokenReadings{
		sentStart(),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("run", &pos, nil), 0),
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	rule := NewDisambiguationPatternRule(
		"D2", "imm run", "en",
		[]*patterns.PatternToken{patterns.Token("run")},
		"", nil, ActionImmunize,
	)
	d := NewXmlRuleDisambiguator([]*DisambiguationPatternRule{rule})
	out := d.Disambiguate(sent)
	found := false
	for _, t := range out.GetTokensWithoutWhitespace() {
		if t.GetToken() == "run" && t.IsImmunized() {
			found = true
		}
	}
	require.True(t, found)
}

func TestDisambiguatedExample(t *testing.T) {
	e := NewDisambiguatedExampleFull("He can can.", "can[can/NN]", "can[can/VB]")
	require.Contains(t, e.String(), "can/NN")
}

func TestDisambiguationPatternRule_ReplacePreservesPreDisambig(t *testing.T) {
	// Java: return new AnalyzedSentence(whTokens, preDisambigTokens) where
	// preDisambigTokens is sentence.getTokens() before executeAction reassignments.
	// REPLACE reassigns a new ATR so pre-disambig keeps the original readings.
	posNN, posVB := "NN", "VB"
	lemma := "run"
	orig := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("run", &posNN, &lemma), 0)
	// also has VB reading
	orig.AddReading(languagetool.NewAnalyzedToken("run", &posVB, &lemma), "tagger")
	toks := []*languagetool.AnalyzedTokenReadings{sentStart(), orig}
	sent := languagetool.NewAnalyzedSentence(toks)
	require.Len(t, orig.GetReadings(), 2)

	rule := NewDisambiguationPatternRule(
		"REP", "replace to VB", "en",
		[]*patterns.PatternToken{patterns.Token("run")},
		"VB", nil, ActionReplace,
	)
	out := rule.Replace(sent)
	require.NotNil(t, out)

	// Working tokens: single VB reading after REPLACE
	var runOut *languagetool.AnalyzedTokenReadings
	for _, t := range out.GetTokensWithoutWhitespace() {
		if t != nil && t.GetToken() == "run" {
			runOut = t
		}
	}
	require.NotNil(t, runOut)
	require.Len(t, runOut.GetReadings(), 1)
	require.Equal(t, "VB", *runOut.GetReadings()[0].GetPOSTag())

	// Pre-disambig still has both original readings (Java preDisambigTokens array).
	var runPre *languagetool.AnalyzedTokenReadings
	for _, t := range out.GetPreDisambigTokensWithoutWhitespace() {
		if t != nil && t.GetToken() == "run" {
			runPre = t
		}
	}
	require.NotNil(t, runPre)
	require.NotSame(t, runOut, runPre)
	require.Len(t, runPre.GetReadings(), 2)
}

func TestDisambiguationPatternRule_IgnoreSpelling(t *testing.T) {
	posNN := "NN"
	toks := []*languagetool.AnalyzedTokenReadings{
		sentStart(),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("foo", &posNN, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("bar", &posNN, nil), 4),
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	rule := NewDisambiguationPatternRule(
		"D3", "ignore foo", "en",
		[]*patterns.PatternToken{patterns.Token("foo")},
		"", nil, ActionIgnoreSpelling,
	)
	out := rule.Replace(sent)
	found := false
	for _, tok := range out.GetTokensWithoutWhitespace() {
		if tok.GetToken() == "foo" && tok.IsIgnoredBySpeller() {
			found = true
		}
	}
	require.True(t, found)
}

func TestDisambiguationPatternRule_AddChunk(t *testing.T) {
	// Java ADDCHUNK requires <wd pos=…/> list matching marker span length.
	posNN := "NN"
	toks := []*languagetool.AnalyzedTokenReadings{
		sentStart(),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("New", &posNN, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("York", &posNN, nil), 4),
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	rule := NewDisambiguationPatternRule(
		"D4", "chunk NP", "en",
		[]*patterns.PatternToken{patterns.Token("New"), patterns.Token("York")},
		"", nil, ActionAddChunk,
	)
	bnp, inp := "B-NP", "I-NP"
	rule.SetNewInterpretations([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("", &bnp, nil),
		languagetool.NewAnalyzedToken("", &inp, nil),
	})
	out := rule.Replace(sent)
	foundB, foundI := false, false
	for _, tok := range out.GetTokensWithoutWhitespace() {
		if tok.GetToken() == "New" {
			for _, c := range tok.GetChunkTags() {
				if c == "B-NP" {
					foundB = true
				}
			}
		}
		if tok.GetToken() == "York" {
			for _, c := range tok.GetChunkTags() {
				if c == "I-NP" {
					foundI = true
				}
			}
		}
	}
	require.True(t, foundB, "expected B-NP chunk on New")
	require.True(t, foundI, "expected I-NP chunk on York")
}

// TestDisambiguationPatternRule_ReplaceLemmaFallbackEmptyOnly ports Java
// DisambiguationPatternRuleReplacer REPLACE (~320–329): after scanning for exact
// POS match lemma, fall back to getAnalyzedToken(0).getLemma() only when the
// collected lemma is empty — not when the matched lemma equals the surface form.
func TestDisambiguationPatternRule_ReplaceLemmaFallbackEmptyOnly(t *testing.T) {
	// Token "a" with dict-like order: first reading mont/V…, later a/P.
	// REPLACE to P must keep lemma "a", not overwrite with token0 "mont".
	posV, posP := "V", "P"
	lemmaMont, lemmaA := "mont", "a"
	orig := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("a", &posV, &lemmaMont), 0)
	orig.AddReading(languagetool.NewAnalyzedToken("a", &posP, &lemmaA), "tagger")
	toks := []*languagetool.AnalyzedTokenReadings{sentStart(), orig}
	sent := languagetool.NewAnalyzedSentence(toks)

	rule := NewDisambiguationPatternRule(
		"PREP_A", "a → P", "br",
		[]*patterns.PatternToken{patterns.Token("a")},
		"P", nil, ActionReplace,
	)
	out := rule.Replace(sent)
	var aOut *languagetool.AnalyzedTokenReadings
	for _, tok := range out.GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == "a" {
			aOut = tok
		}
	}
	require.NotNil(t, aOut)
	require.Len(t, aOut.GetReadings(), 1)
	require.Equal(t, "P", *aOut.GetReadings()[0].GetPOSTag())
	require.NotNil(t, aOut.GetReadings()[0].GetLemma())
	require.Equal(t, "a", *aOut.GetReadings()[0].GetLemma(),
		"REPLACE must keep POS-matched lemma, not surface==lemma fallback to token0")

	// No matching POS: lemma stays empty → fallback to token0 lemma (may be null).
	posX := "X"
	untagged := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("XXI", nil, nil), 0)
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{sentStart(), untagged})
	rule2 := NewDisambiguationPatternRule(
		"XXI", "roman", "br",
		[]*patterns.PatternToken{patterns.Token("XXI")},
		posX, nil, ActionReplace,
	)
	out2 := rule2.Replace(sent2)
	var xOut *languagetool.AnalyzedTokenReadings
	for _, tok := range out2.GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == "XXI" {
			xOut = tok
		}
	}
	require.NotNil(t, xOut)
	require.Len(t, xOut.GetReadings(), 1)
	require.Equal(t, "X", *xOut.GetReadings()[0].GetPOSTag())
	require.Nil(t, xOut.GetReadings()[0].GetLemma(),
		"empty-lemma fallback uses token0 lemma (null here), not surface as lemma")
}
