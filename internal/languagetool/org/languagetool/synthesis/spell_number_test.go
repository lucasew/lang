package synthesis

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestBaseSynthesizer_SpellNumberTags(t *testing.T) {
	s := NewBaseSynthesizer("en", nil)
	s.LoadNumberSpellersFromDir("") // walk-up inspiration
	if s.NumberSpeller == nil {
		t.Skip("en.sor not found")
	}
	// Java EnglishSynthesizerTest
	tok := languagetool.NewAnalyzedToken("12", nil, nil)
	got, err := s.Synthesize(tok, SpellNumberTag)
	require.NoError(t, err)
	require.Equal(t, []string{"twelve"}, got)

	tok2 := languagetool.NewAnalyzedToken("1243", nil, nil)
	got, err = s.Synthesize(tok2, SpellNumberTag)
	require.NoError(t, err)
	require.Equal(t, []string{"one thousand two hundred forty-three"}, got)

	if s.RomanNumberer != nil {
		tok3 := languagetool.NewAnalyzedToken("12", nil, nil)
		got, err = s.Synthesize(tok3, SpellNumberRomanTag)
		require.NoError(t, err)
		require.Equal(t, []string{"XII"}, got)
	}
}

func TestBaseSynthesizer_GetTargetPosTag_Last(t *testing.T) {
	// Java BaseSynthesizer.getTargetPosTag returns last element
	s := NewBaseSynthesizer("en", nil)
	require.Equal(t, "X", s.GetTargetPosTag(nil, "X"))
	require.Equal(t, "B", s.GetTargetPosTag([]string{"A", "B"}, "X"))
}

func TestBaseSynthesizer_SpellNumberFeminine_Portuguese(t *testing.T) {
	s := NewBaseSynthesizer("pt", nil)
	s.SorFileName = "/pt/pt.sor"
	s.LoadNumberSpellersFromDir("")
	if s.NumberSpeller == nil {
		t.Skip("pt.sor not found")
	}
	// Java: _spell_number_:feminine → getSpelledNumber("feminine " + token)
	tok := languagetool.NewAnalyzedToken("2", nil, nil)
	got, err := s.Synthesize(tok, SpellNumberFeminineTag)
	require.NoError(t, err)
	require.Len(t, got, 1)
	// feminine 2 → duas (via == feminine == macros)
	require.Equal(t, "duas", got[0], "feminine spell of 2")
	// masculine default
	gotM, err := s.Synthesize(tok, SpellNumberTag)
	require.NoError(t, err)
	require.Equal(t, []string{"dois"}, gotM)
}
