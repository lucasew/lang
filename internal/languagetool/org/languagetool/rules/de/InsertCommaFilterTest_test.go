package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/InsertCommaFilterTest.java
//
// Java:
//   JLanguageTool lt = new JLanguageTool(Languages.getLanguageForShortCode("de"));
//   runFilter(input, patternTokenPos, atrFromPos, atrToPos, fromPos, toPos, origRepl, newRepl, lt);
//   InsertCommaFilter uses GermanTagger.INSTANCE.tag on suggestion parts.

import (
	"fmt"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// TestInsertCommaFilter_Filter ports InsertCommaFilterTest.testFilter.
func TestInsertCommaFilter_Filter(t *testing.T) {
	// Java: new JLanguageTool(Languages.getLanguageForShortCode("de"))
	lt := languagetool.NewJLanguageTool("de")
	// Wire tagger for GetAnalyzedSentence POS (Java createDefaultTagger → GermanTagger.INSTANCE).
	require.True(t, WireGermanTagWord(lt), "GermanTagger.INSTANCE resources required")
	WireGermanTaggerDefaults()
	require.NotNil(t, DiscoveredGermanTagger(), "GermanTagger.INSTANCE resources required")

	runFilter(t, "Das Auto, das am Straßenrand steht parkt im Halteverbot.", 7,
		7, 8, 29, 40, "steht parkt", "[steht, parkt]", lt)
	runFilter(t, "Hoffe bei euch ist alles gut.", 1,
		1, 4, 0, 7, "Hoffe bei euch ist", "[Hoffe, bei euch ist]", lt)
}

// runFilter ports InsertCommaFilterTest.runFilter.
func runFilter(t *testing.T, input string, patternTokenPos, atrFromPos, atrToPos, fromPos, toPos int,
	origRepl, newRepl string, lt *languagetool.JLanguageTool) {
	t.Helper()
	args := map[string]string{}
	filter := NewInsertCommaFilter()
	sentence := lt.GetAnalyzedSentence(input)
	require.NotNil(t, sentence, "GetAnalyzedSentence(%q)", input)
	match := rules.NewRuleMatch(rules.NewFakeRule("I"), sentence, fromPos, toPos, "fake msg")
	match.SetSuggestedReplacement(origRepl)
	// Java: AnalyzedTokenReadings[] atr = { tokensWithoutWhitespace[atrFromPos], tokensWithoutWhitespace[atrToPos] }
	toks := sentence.GetTokensWithoutWhitespace()
	require.Greater(t, len(toks), atrFromPos)
	require.Greater(t, len(toks), atrToPos)
	atr := []*languagetool.AnalyzedTokenReadings{toks[atrFromPos], toks[atrToPos]}
	matchTmp := filter.AcceptRuleMatch(match, args, patternTokenPos, atr, nil)
	require.NotNil(t, matchTmp)
	// Java: assertThat(matchTmp.getSuggestedReplacements().toString(), is(newRepl));
	got := fmt.Sprint(matchTmp.GetSuggestedReplacements())
	require.Equal(t, newRepl, got)
}
