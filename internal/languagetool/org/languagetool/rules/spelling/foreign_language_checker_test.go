package spelling

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language/identifier"
	"github.com/stretchr/testify/require"
)

// saveRestoreIdentifier clears LanguageIdentifierService for the test and cleans up after.
func saveRestoreIdentifier(t *testing.T) {
	t.Helper()
	identifier.Instance.Clear()
	t.Cleanup(func() {
		identifier.Instance.Clear()
	})
}

func TestForeignLanguageChecker_ShortSentence_Empty(t *testing.T) {
	saveRestoreIdentifier(t)
	// sentenceLength < MIN_SENTENCE_THRESHOLD (3) → empty regardless of ratio / identifier.
	id := identifier.NewMapLanguageIdentifier(100, func(text string, pref []string) map[string]float64 {
		return map[string]float64{"de": 0.9}
	})
	identifier.Instance.SetSimple(id)
	c := NewForeignLanguageChecker("en", "hi", 2, []string{"en", "de"})
	require.Empty(t, c.Check(10))
}

func TestForeignLanguageChecker_LowErrorRatio_Empty(t *testing.T) {
	saveRestoreIdentifier(t)
	id := identifier.NewMapLanguageIdentifier(100, func(text string, pref []string) map[string]float64 {
		return map[string]float64{"de": 0.9}
	})
	identifier.Instance.SetSimple(id)
	// 1/10 = 0.1 < ERROR_THRESHOLD 0.45
	c := NewForeignLanguageChecker("en", "hello world there more words", 10, []string{"en", "de"})
	require.Empty(t, c.Check(1))
}

func TestForeignLanguageChecker_HighRatio_NilIdentifier_Empty(t *testing.T) {
	saveRestoreIdentifier(t)
	// GetInitialized() == nil → fall through to empty map.
	require.Nil(t, identifier.Instance.GetInitialized())
	c := NewForeignLanguageChecker("en", "hello world there more words", 10, []string{"en", "de"})
	require.Empty(t, c.Check(5))
}

func TestForeignLanguageChecker_HighRatio_SameLangFirst_NoForeign(t *testing.T) {
	saveRestoreIdentifier(t)
	id := identifier.NewMapLanguageIdentifier(100, func(text string, pref []string) map[string]float64 {
		return map[string]float64{"en": 0.95, "de": 0.4}
	})
	identifier.Instance.SetSimple(id)
	c := NewForeignLanguageChecker("en", "hello world there more words", 10, []string{"en", "de"})
	m := c.Check(5)
	require.Equal(t, map[string]float32{NoForeignLangDetected: 0.99}, m)
}

func TestForeignLanguageChecker_HighRatio_OtherLangScores(t *testing.T) {
	saveRestoreIdentifier(t)
	id := identifier.NewMapLanguageIdentifier(100, func(text string, pref []string) map[string]float64 {
		return map[string]float64{"de": 0.9, "fr": 0.4}
	})
	identifier.Instance.SetSimple(id)
	c := NewForeignLanguageChecker("en", "hello world there more words", 10, []string{"en", "de", "fr"})
	m := c.Check(5)
	require.Equal(t, float32(0.9), m["de"])
	require.Equal(t, float32(0.4), m["fr"])
	require.NotContains(t, m, NoForeignLangDetected)
	require.NotContains(t, m, "en")
}

func TestForeignLanguageChecker_HighRatio_EmptyScores_NoForeign(t *testing.T) {
	saveRestoreIdentifier(t)
	// MapLanguageIdentifier with Score returning empty map → Scores empty → NO_FOREIGN_LANG_DETECTED.
	id := identifier.NewMapLanguageIdentifier(100, func(text string, pref []string) map[string]float64 {
		return map[string]float64{}
	})
	identifier.Instance.SetSimple(id)
	c := NewForeignLanguageChecker("en", "hello world there more words", 10, []string{"en", "de"})
	m := c.Check(5)
	require.Equal(t, map[string]float32{NoForeignLangDetected: 0.99}, m)
}

func TestForeignLanguageChecker_HighRatio_NilScoreFunc_EmptyScores_NoForeign(t *testing.T) {
	saveRestoreIdentifier(t)
	// Score == nil → Scores returns nil/empty → NO_FOREIGN_LANG_DETECTED.
	id := identifier.NewMapLanguageIdentifier(100, nil)
	identifier.Instance.SetSimple(id)
	c := NewForeignLanguageChecker("en", "hello world there more words", 10, []string{"en", "de"})
	m := c.Check(5)
	require.Equal(t, map[string]float32{NoForeignLangDetected: 0.99}, m)
}

// Preferred list is copied (Java Collections.unmodifiableList of a snapshot).
func TestForeignLanguageChecker_PreferredCopied(t *testing.T) {
	saveRestoreIdentifier(t)
	pref := []string{"en", "de"}
	c := NewForeignLanguageChecker("en", "x", 10, pref)
	pref[0] = "fr"
	// Identifier receives preferredLanguages from the checker (original copy).
	id := identifier.NewMapLanguageIdentifier(100, func(text string, preferred []string) map[string]float64 {
		require.Equal(t, []string{"en", "de"}, preferred)
		return map[string]float64{"de": 0.8}
	})
	identifier.Instance.SetSimple(id)
	m := c.Check(5)
	require.Equal(t, float32(0.8), m["de"])
}
