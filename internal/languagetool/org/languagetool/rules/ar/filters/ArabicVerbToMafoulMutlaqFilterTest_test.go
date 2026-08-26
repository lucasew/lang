package filters

// Twin of ArabicVerbToMasdarFilterTest (Java king) — class misnamed in Java;
// constructs new ArabicVerbToMafoulMutlaqFilter() and exercises acceptRuleMatch.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	ar_tag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ar"
	"github.com/stretchr/testify/require"
)

// TestArabicVerbToMafoulMutlaqFilter_Filter twins ArabicVerbToMasdarFilterTest.testFilter.
// Java assertSuggestion("يعمل", "يعمل عملًا جيدًا"): tokens يعمل, بأسلوب, جيد; args verb/adj.
func TestArabicVerbToMafoulMutlaqFilter_Filter(t *testing.T) {
	assertMafoulMutlaqSuggestion(t, "يعمل", "يعمل عملًا جيدًا")
}

func assertMafoulMutlaqSuggestion(t *testing.T, word, expectedSuggestion string) {
	t.Helper()
	ar_tag.EnsureDefaultArabicTagger()
	require.NotNil(t, ar_tag.DefaultArabicTagger)
	require.NotEmpty(t, ar_tag.ArabicPOSDictPath(), "arabic.dict required for Java twin")

	f := NewArabicVerbToMafoulMutlaqFilter()
	require.NotNil(t, f.synthesizer, "ArabicSynthesizer.INSTANCE required for Java twin")
	require.NotNil(t, f.tagger)

	word2 := "بأسلوب"
	word3 := "جيد"
	args := map[string]string{
		"verb": word,
		"adj":  word3,
	}
	patternTokens := ar_tag.DefaultArabicTagger.Tag([]string{word, word2, word3})
	require.Len(t, patternTokens, 3)

	match := rules.NewRuleMatch(nil, nil, 0, 10, "message")
	ruleMatch := f.AcceptRuleMatch(match, args, -1, patternTokens, nil)

	require.NotNil(t, ruleMatch)
	sugs := ruleMatch.GetSuggestedReplacements()
	require.Contains(t, sugs, expectedSuggestion, "word=%s suggestions=%v", word, sugs)
}

// TestArabicVerbToMafoulMutlaqFilter_LoadFromPath twins loadFromPath resource resolve
// (Java test defines loadFromPath; map must load official /ar/arabic_verb_masdar.txt).
func TestArabicVerbToMafoulMutlaqFilter_LoadFromPath(t *testing.T) {
	m := LoadVerbMasdarFromPath(ArabicVerbMasdarFileName)
	require.NotNil(t, m)
	require.NotEmpty(t, m, "Java loadWords(/ar/arabic_verb_masdar.txt) must resolve")
	// Known verb→masdar rows from official file (exact keys, last-wins on duplicate keys)
	// File has عَمِلَ=عمل early then later عَمِلَ=عملة|عمل — SimpleReplace last put wins.
	require.Equal(t, []string{"عملة", "عمل"}, m["عَمِلَ"])
	require.Equal(t, []string{"إعمال"}, m["أَعْمَلَ"])
	require.Equal(t, []string{"أكل"}, m["أَكَلَ"])
	require.Equal(t, []string{"سؤال"}, m["سَأَلَ"])
	// No tashkeel-stripped invent key
	require.Nil(t, m["عمل"])
}
