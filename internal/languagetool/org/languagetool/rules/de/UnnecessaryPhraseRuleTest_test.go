package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/UnnecessaryPhraseRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUnnecessaryPhraseRule_Rule(t *testing.T) {
	// Java: UnnecessaryPhraseRule(..., new UserConfig()) → default min 8 per-mill.
	// Do NOT set MinPercent=0: Java (minPercent==0 && !withoutDirectSpeech) still counts
	// tokens inside direct speech, so exclusion only works with the production default.
	rule := NewUnnecessaryPhraseRule(nil)
	require.Equal(t, 8, rule.MinPercent)
	require.True(t, rule.ExcludeDirectSpeech)

	// Java: 4 phrases — "In diesem Zusammenhang", "im Allgemeinen", "voll und ganz", "mehr oder weniger"
	// ("zumindest" is not in the phrase list)
	ms := rule.Match(languagetool.AnalyzePlain(
		"In diesem Zusammenhang ist es im Allgemeinen voll und ganz zumindest mehr oder weniger sinnvoll."))
	require.Equal(t, 4, len(ms), "four unnecessary phrases")

	// Java: exclude direct speech with German low/high quotes „…“
	ms = rule.Match(languagetool.AnalyzePlain(
		"„In diesem Zusammenhang ist es im Allgemeinen voll und ganz zumindest mehr oder weniger sinnvoll.“"))
	require.Equal(t, 0, len(ms), "direct speech excluded")

	// no phrase
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Es ist weniger sinnvoll."))))

	// class example pair
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Das ist allem Anschein nach eine Phrase."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Das ist eine Phrase."))))
}

func TestUnnecessaryPhraseRule_Meta(t *testing.T) {
	r := NewUnnecessaryPhraseRule(nil)
	require.Equal(t, "UNNECESSARY_PHRASES_DE", r.GetID())
	require.Equal(t, 8, r.MinPercent)
	require.True(t, r.ExcludeDirectSpeech)
	require.True(t, r.IsDefaultOff())
	require.NotEmpty(t, r.GetIncorrectExamples())
}

// MinPercent==0 disables direct-speech skip (Java twin of the compound guard).
func TestUnnecessaryPhraseRule_MinZeroStillSeesDirectSpeech(t *testing.T) {
	rule := NewUnnecessaryPhraseRule(nil)
	rule.MinPercent = 0
	// With min 0, phrases inside „…“ still fire (unless WithoutDirectSpeech).
	ms := rule.Match(languagetool.AnalyzePlain(
		"„In diesem Zusammenhang ist es im Allgemeinen voll und ganz zumindest mehr oder weniger sinnvoll.“"))
	// first "In diesem Zusammenhang" may miss if nToken!=1 (quote is token 1) so capitalization
	// keeps "In" ≠ "in"; remaining three still match.
	require.GreaterOrEqual(t, len(ms), 3)
}
