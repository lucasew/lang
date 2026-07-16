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
