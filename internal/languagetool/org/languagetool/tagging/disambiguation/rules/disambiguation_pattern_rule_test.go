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
		"B-NP", nil, ActionAddChunk,
	)
	out := rule.Replace(sent)
	found := false
	for _, tok := range out.GetTokensWithoutWhitespace() {
		if tok.GetToken() == "New" {
			for _, c := range tok.GetChunkTags() {
				if c == "B-NP" {
					found = true
				}
			}
		}
	}
	require.True(t, found, "expected B-NP chunk on New")
}
