package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func ptrAV(s string) *string { return &s }

func atrAV(token, pos, lemma string, start int) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(token, ptrAV(pos), ptrAV(lemma)), start)
}

func sentenceAV(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	start := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("", ptrAV(languagetool.SentenceStartTagName), nil))
	return languagetool.NewAnalyzedSentence(append([]*languagetool.AnalyzedTokenReadings{start}, toks...))
}

func TestAdjustVerbSuggestionsFilter_Suggest(t *testing.T) {
	f := NewAdjustVerbSuggestionsFilter()
	got := f.Suggest(VerbSuggestionContext{
		PronounsStr:   "em",
		VerbStr:       "vaig",
		WholeOriginal: "em vaig",
	}, "replaceEmEn")
	require.Equal(t, []string{"en vaig"}, got)

	got = f.Suggest(VerbSuggestionContext{
		PronounsStr:            "",
		VerbStr:                "menja",
		FirstVerbPersonaNumber: "3S",
		WholeOriginal:          "menja",
	}, "addPronounReflexive")
	require.NotEmpty(t, got)
}

func TestAdjustVerbRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ca.AdjustVerbSuggestionsFilter"))
}

func TestCatalanGetTargetPosTag(t *testing.T) {
	// 3S preferred over 1S → last after sort
	got := catalanGetTargetPosTag([]string{"VMIP1S00", "VMIP3S00"}, "")
	require.Equal(t, "VMIP3S00", got)
	require.Equal(t, "fallback", catalanGetTargetPosTag(nil, "fallback"))
}

// em vaig + suggestion "menjar-se" → addPronounReflexive on menjar
func TestAdjustVerbSuggestionsFilter_AcceptReflexive(t *testing.T) {
	f := NewAdjustVerbSuggestionsFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, postag string) []string {
		lem := ""
		if tok.GetLemma() != nil {
			lem = *tok.GetLemma()
		}
		if lem == "menjar" && postag == "VMIP1S00" {
			return []string{"menjo"}
		}
		return nil
	}
	// tokens: em (P0) + vaig (V)
	em := atrAV("em", "P01CN000", "jo", 0)
	em.SetWhitespaceBefore(true)
	vaig := atrAV("vaig", "VMIP1S00", "anar", 3)
	vaig.SetWhitespaceBefore(true)
	sent := sentenceAV(em, vaig)
	m := rules.NewRuleMatch(nil, sent, 0, 7, "msg")
	m.SetSuggestedReplacements([]string{"menjar-se"})
	out := f.AcceptRuleMatch(m, map[string]string{
		"actions": "removePronounReflexive", // overridden by -se suffix on lemma
	}, 0, nil, nil)
	require.NotNil(t, out)
	sugs := out.GetSuggestedReplacements()
	require.NotEmpty(t, sugs)
	// should contain reflexive form of menjo
	require.Contains(t, sugs[0], "menjo")
}

func TestAdjustVerbSuggestionsFilter_NoSynth(t *testing.T) {
	f := NewAdjustVerbSuggestionsFilter()
	require.Nil(t, f.AcceptRuleMatch(nil, nil, 0, nil, nil))
}

func TestAdjustVerbSuggestionsFilter_EmptySuggestions(t *testing.T) {
	f := NewAdjustVerbSuggestionsFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, postag string) []string {
		return []string{"x"}
	}
	vaig := atrAV("vaig", "VMIP1S00", "anar", 0)
	sent := sentenceAV(vaig)
	m := rules.NewRuleMatch(nil, sent, 0, 4, "msg")
	// no suggested replacements → empty loop → nil
	require.Nil(t, f.AcceptRuleMatch(m, nil, 0, nil, nil))
}

func TestAnyChangeVowelConsonant(t *testing.T) {
	require.True(t, anyChangeVowelConsonant("amic", []string{"casa"}))
	require.False(t, anyChangeVowelConsonant("casa", []string{"cosa"}))
}
