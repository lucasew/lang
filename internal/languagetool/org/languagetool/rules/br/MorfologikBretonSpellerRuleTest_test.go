package br

// Twin of org.languagetool.rules.br.MorfologikBretonSpellerRuleTest
// Official resource: /br/hunspell/br_FR.dict (+ spelling.txt via InitSpellersFromGetters).
// No map-inject / ClearMultiSpellers invent — real binary dict only.
//
// Java setup:
//   rule = new MorfologikBretonSpellerRule(TestTools.getMessages("br"), new Breton(), null, emptyList());
//   lt   = new JLanguageTool(new Breton()); // BretonWordTokenizer
// Match positions are UTF-16 (Java RuleMatch fromPos/toPos).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	brtok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/br"
	"github.com/stretchr/testify/require"
)

// requireBretonSpellerDict skips when official br_FR.dict is not discoverable (no invent).
func requireBretonSpellerDict(t *testing.T) *MorfologikBretonSpellerRule {
	t.Helper()
	if morfologik.DiscoverLanguageDict(MorfologikBretonSpellerRuleDict) == "" {
		t.Skip("br_FR.dict not discoverable (inspiration/.../br/hunspell)")
	}
	r := NewMorfologikBretonSpellerRule()
	require.NotNil(t, r)
	require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	return r
}

// analyzeBR ports JLanguageTool(new Breton()).getAnalyzedSentence: BretonWordTokenizer.
func analyzeBR(text string) *languagetool.AnalyzedSentence {
	return languagetool.AnalyzeWithTokenizer(text, brtok.NewBretonWordTokenizer())
}

func matchBR(t *testing.T, r *MorfologikBretonSpellerRule, text string) []*rules.RuleMatch {
	t.Helper()
	ms, err := r.Match(analyzeBR(text))
	require.NoError(t, err)
	return ms
}

// Port of MorfologikBretonSpellerRuleTest.testMorfologikSpeller
func TestMorfologikBretonSpellerRule_MorfologikSpeller(t *testing.T) {
	r := requireBretonSpellerDict(t)
	require.Equal(t, MorfologikBretonSpellerRuleID, r.GetID())
	require.Equal(t, MorfologikBretonSpellerRuleDict, r.GetFileName())
	require.Equal(t, "MORFOLOGIK_RULE_BR_FR", r.GetID())
	require.Equal(t, "/br/hunspell/br_FR.dict", r.GetFileName())
	// Java ctor: setIgnoreTaggedWords()
	require.True(t, r.IgnoreTaggedWords)
	// Java tokenizingPattern(): Pattern.compile("-")
	require.NotNil(t, r.TokenizingPattern)
	require.True(t, r.TokenizingPattern.MatchString("-"))

	// correct sentences:
	require.Len(t, matchBR(t, r, "Penaos emañ kont ganit?"), 0)

	require.Len(t, matchBR(t, r, "C'hwerc'h merc'h gwerc'h war c'hwerc'h marc'h kalloc'h"), 0)
	require.Len(t, matchBR(t, r, "C’hwerc’h merc’h gwerc‘h war c‘hwerc‘h marc'h kalloc‘h"), 0)

	// words with hyphens are tokenized internally...
	require.Len(t, matchBR(t, r, "Evel-just"), 0)
	require.Len(t, matchBR(t, r, "Barrek-tre eo LanguageTool"), 0)

	require.Len(t, matchBR(t, r, "C'hwerc'h merc'h gwerc'h war c'hwerc'h marc'h kalloc'h"), 0)
	require.Len(t, matchBR(t, r, "C’hwerc’h merc’h gwerc‘h war c‘hwerc‘h marc'h kalloc‘h"), 0)
	require.Len(t, matchBR(t, r, "Evel-just"), 0)
	require.Len(t, matchBR(t, r, "Evel-juste"), 1)
	require.Len(t, matchBR(t, r, "Barrek-tre eo LanguageTool"), 0)

	// test for "LanguageTool":
	require.Len(t, matchBR(t, r, "LanguageTool!"), 0)
	require.Len(t, matchBR(t, r, ","), 0)
	require.Len(t, matchBR(t, r, "123454"), 0)

	// incorrect sentences:
	require.Len(t, matchBR(t, r, "Evel-juste"), 1)

	matches := matchBR(t, r, "Evel-juste")
	// check match positions:
	require.Len(t, matches, 1)
	require.Equal(t, 5, matches[0].GetFromPos())
	require.Equal(t, 10, matches[0].GetToPos())

	matches = matchBR(t, r, "C’hreizhig-don")
	require.Len(t, matches, 1)
	// check match positions:
	require.Len(t, matches, 1)
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 10, matches[0].GetToPos())

	require.Len(t, matchBR(t, r, "aõh"), 1)
	require.Len(t, matchBR(t, r, "a"), 0)
}
