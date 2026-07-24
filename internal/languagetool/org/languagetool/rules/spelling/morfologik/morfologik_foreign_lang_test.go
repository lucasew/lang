package morfologik

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/stretchr/testify/require"
)

// Twin of MorfologikSpellerRule.match ForeignLanguageChecker arm:
// preferredLanguages ≥ 2 + high error ratio → setNewLanguageMatches on first match.
func TestMatch_ForeignLanguageChecker_SetsNewLanguageMatches(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	// Only a few correct English words; rest misspelled so errorRatio ≥ 0.45.
	for _, w := range []string{"the", "and", "a"} {
		sp.AddWord(w)
	}
	r := NewMorfologikSpellerRule("TEST", "en", "/xx.dict", sp)
	uc := languagetool.NewUserConfig()
	uc.SetPreferredLanguagesList([]string{"en", "de", "fr"})
	r.SetUserConfig(uc)
	// Inject identifier: claim German (Java LanguageIdentifier path).
	r.ForeignDetect = func(sentence string, preferred []string, maxResults int) []spelling.DetectedLanguageScore {
		return []spelling.DetectedLanguageScore{
			{ShortCode: "de", Confidence: 0.91, Source: "test"},
			{ShortCode: "fr", Confidence: 0.4, Source: "test"},
		}
	}

	// ≥ 3 non-nonWord tokens; most misspelled.
	// "xxx yyy zzz www the" → 5 content words-ish; many matches.
	ms, err := r.Match(languagetool.AnalyzePlain("xxx yyy zzz www the"))
	require.NoError(t, err)
	require.NotEmpty(t, ms, "expected spelling matches for unknown words")
	// First match carries foreign language scores (Java ruleMatches.get(0).setNewLanguageMatches).
	nl := ms[0].GetNewLanguageMatches()
	require.Contains(t, nl, "de")
	require.InDelta(t, 0.91, float64(nl["de"]), 0.001)
	require.NotContains(t, nl, spelling.NoForeignLangDetected)
}

func TestMatch_ForeignLanguageChecker_SameLang_NoNewLanguageMatches(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	// Non-empty map so IsMisspelled flags unknowns (empty Words = fail-closed).
	sp.AddWord("the")
	r := NewMorfologikSpellerRule("TEST", "en", "/xx.dict", sp)
	uc := languagetool.NewUserConfig()
	uc.SetPreferredLanguagesList([]string{"en", "de"})
	r.SetUserConfig(uc)
	r.ForeignDetect = func(sentence string, preferred []string, maxResults int) []spelling.DetectedLanguageScore {
		// Top hit is same language as the rule → NO_FOREIGN_LANG_DETECTED (no set).
		return []spelling.DetectedLanguageScore{
			{ShortCode: "en", Confidence: 0.95, Source: "test"},
		}
	}
	ms, err := r.Match(languagetool.AnalyzePlain("xxx yyy zzz www aaa"))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	require.Empty(t, ms[0].GetNewLanguageMatches())
}

func TestPreferredLanguagesActive(t *testing.T) {
	require.False(t, preferredLanguagesActive(nil))
	require.False(t, preferredLanguagesActive([]string{""}))
	require.False(t, preferredLanguagesActive([]string{"en"}))
	require.True(t, preferredLanguagesActive([]string{"en", "de"}))
	require.True(t, preferredLanguagesActive([]string{" en ", "de"}))
}

func TestForeignSentenceLength(t *testing.T) {
	// AnalyzePlain: SENT_START + words + punctuation may be non-word.
	s := languagetool.AnalyzePlain("one two three")
	// non-nonWord count - 1 (Java). SENT_START may or may not count as non-word.
	n := foreignSentenceLength(s)
	// Sanity: positive for multi-word sentence.
	require.Greater(t, n, int64(0), "text=%q tokens=%d", s.GetText(), len(s.GetTokensWithoutWhitespace()))
}
