package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglishHasNGramFalseFriendRule(t *testing.T) {
	// Java English.hasNGramFalseFriendRule — de, fr, es, nl only
	require.False(t, EnglishHasNGramFalseFriendRule(""))
	require.True(t, EnglishHasNGramFalseFriendRule("de"))
	require.True(t, EnglishHasNGramFalseFriendRule("fr"))
	require.True(t, EnglishHasNGramFalseFriendRule("es"))
	require.True(t, EnglishHasNGramFalseFriendRule("nl"))
	require.True(t, EnglishHasNGramFalseFriendRule("de-DE"))
	require.True(t, EnglishHasNGramFalseFriendRule("es-MX"))
	require.False(t, EnglishHasNGramFalseFriendRule("it"))
	require.False(t, EnglishHasNGramFalseFriendRule("pt"))
	require.False(t, EnglishHasNGramFalseFriendRule("en"))
	require.False(t, EnglishHasNGramFalseFriendRule("ru"))
}

func TestEnglishVariant_GetMaintainedState(t *testing.T) {
	require.Equal(t, languagetool.ActivelyMaintained, AmericanEnglish.GetMaintainedState())
	require.Equal(t, languagetool.ActivelyMaintained, BritishEnglish.GetMaintainedState())
}

func TestRomanceVariants_GetMaintainedState(t *testing.T) {
	require.Equal(t, languagetool.ActivelyMaintained, FrenchFrance.GetMaintainedState())
	require.Equal(t, languagetool.ActivelyMaintained, SpanishSpain.GetMaintainedState())
	require.Equal(t, languagetool.ActivelyMaintained, BrazilianPortuguese.GetMaintainedState())
}

func TestSmallLang_GetMaintainedState(t *testing.T) {
	// Java overrides to ActivelyMaintained
	require.Equal(t, languagetool.ActivelyMaintained, Swedish.GetMaintainedState())
	require.Equal(t, languagetool.ActivelyMaintained, Greek.GetMaintainedState())
	require.Equal(t, languagetool.ActivelyMaintained, Irish.GetMaintainedState())
	require.Equal(t, languagetool.ActivelyMaintained, Ukrainian.GetMaintainedState())
	require.Equal(t, languagetool.ActivelyMaintained, Esperanto.GetMaintainedState())
	require.Equal(t, languagetool.ActivelyMaintained, Breton.GetMaintainedState())
	require.Equal(t, languagetool.ActivelyMaintained, CrimeanTatar.GetMaintainedState())
	// Java default LookingForNewMaintainer
	require.Equal(t, languagetool.LookingForNewMaintainer, Galician.GetMaintainedState())
	require.Equal(t, languagetool.LookingForNewMaintainer, Slovak.GetMaintainedState())
	require.Equal(t, languagetool.LookingForNewMaintainer, Danish.GetMaintainedState())
	require.Equal(t, languagetool.LookingForNewMaintainer, Japanese.GetMaintainedState())
	require.Equal(t, languagetool.LookingForNewMaintainer, Romanian.GetMaintainedState())
}

func TestMoreVariants_GetMaintainedState(t *testing.T) {
	require.Equal(t, languagetool.ActivelyMaintained, DutchNetherlands.GetMaintainedState())
	require.Equal(t, languagetool.ActivelyMaintained, BelgianDutch.GetMaintainedState())
	require.Equal(t, languagetool.ActivelyMaintained, Catalan.GetMaintainedState())
	require.Equal(t, languagetool.ActivelyMaintained, ValencianCatalan.GetMaintainedState())
	require.Equal(t, languagetool.ActivelyMaintained, Italian.GetMaintainedState())
	require.Equal(t, languagetool.ActivelyMaintained, Polish.GetMaintainedState())
	require.Equal(t, languagetool.ActivelyMaintained, Russian.GetMaintainedState())
	require.Equal(t, languagetool.ActivelyMaintained, Arabic.GetMaintainedState())
	require.Equal(t, languagetool.LookingForNewMaintainer, DefaultSerbian.GetMaintainedState())
}
