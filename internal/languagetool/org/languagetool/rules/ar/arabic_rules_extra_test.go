package ar

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/hunspell"
	"github.com/stretchr/testify/require"
)

func TestArabicHunspellTashkeel(t *testing.T) {
	dict := hunspell.NewMapHunspellDictionary([]string{"كتب", "كتاب"})
	r := NewArabicHunspellSpellerRule(dict)
	require.Equal(t, ArabicHunspellRuleID, r.GetID())
	require.Equal(t, ArabicHunspellDictPath, r.GetDictFilenameInResources("ar"))
	require.False(t, r.IsLatinScript())
	require.False(t, r.IsMisspelledStripped("كتب"))
	require.False(t, r.IsMisspelledStripped("كَتَبَ")) // tashkeel stripped → كتب
	require.True(t, r.IsMisspelledStripped("xyzzy"))

	toks := TokenizeArabicSpellText("مرحبا، عالم!")
	require.Contains(t, toks, "مرحبا")
	require.Contains(t, toks, "عالم")

	// sentence match
	pos := "N"
	readings := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("كَتب", &pos, nil),
	}, 0)
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{readings})
	matches, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, matches)
}

func TestArabicConfusionRuleConstruct(t *testing.T) {
	r := NewArabicConfusionProbabilityRule(nil)
	require.NotNil(t, r)
	require.NotNil(t, r.ConfusionProbabilityRule)
}
