package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func withUS(words ...string) *MorfologikVariantSpellerRule {
	r := NewMorfologikAmericanSpellerRule()
	sp := morfologik.NewMorfologikSpeller(AmericanSpellerDict, 1)
	for _, w := range words {
		sp.AddWord(w)
	}
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	return r
}

func TestMorfologikAmericanSpellerRule_SuggestionForMisspelledHyphenatedWords(t *testing.T) {
	r := withUS("well-known")
	require.False(t, r.Speller.IsMisspelled("well-known"))
	require.True(t, r.Speller.IsMisspelled("wel-known"))
}

func TestMorfologikAmericanSpellerRule_NamedEntityIgnore(t *testing.T) {
	r := withUS("Microsoft")
	require.True(t, r.AcceptWord("Microsoft"))
}

func TestMorfologikAmericanSpellerRule_Suggestions(t *testing.T) {
	r := withUS("color")
	r.Speller.Suggestions["colour"] = []string{"color"}
	require.Equal(t, []string{"color"}, r.Speller.FindReplacements("colour"))
}

func TestMorfologikAmericanSpellerRule_SuggestionForIrregularWords(t *testing.T) {
	r := withUS("went", "go")
	ms, err := r.Match(languagetool.AnalyzePlain("went gose"))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
}
