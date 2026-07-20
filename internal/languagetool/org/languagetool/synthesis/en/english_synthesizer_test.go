package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

// injectSuggestAorAn for unit tests without rules/en init cycle.
func injectSuggestAorAn(s *EnglishSynthesizer) {
	s.SuggestAorAn = func(w string) string {
		// Minimal twin of SuggestAorAn using same cases as EnglishSynthesizerTest.
		lw := tools.LowercaseFirstCharIfCapitalized(w)
		switch stringsToLower(lw) {
		case "hour", "honest", "heir", "apple":
			return "an " + lw
		case "university", "hexagon", "string":
			return "a " + lw
		default:
			// rough: vowel → an
			if lw != "" {
				switch lw[0] {
				case 'a', 'e', 'i', 'o':
					return "an " + lw
				}
			}
			return "a " + lw
		}
	}
}

func stringsToLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func TestEnglishSynthesizerDeterminers(t *testing.T) {
	s := NewEnglishSynthesizer(nil)
	injectSuggestAorAn(s)
	tok := languagetool.NewAnalyzedToken("apple", nil, strp("apple"))
	got, err := s.Synthesize(tok, AddIndDeterminer)
	require.NoError(t, err)
	require.Equal(t, []string{"an apple"}, got)
	got, err = s.Synthesize(tok, AddDeterminer)
	require.NoError(t, err)
	// Java: {aOrAn, "the " + lowercaseFirstCharIfCapitalized(token)}
	require.Equal(t, []string{"an apple", "the apple"}, got)
}

func TestEnglishSynthesizer_IsException(t *testing.T) {
	s := NewEnglishSynthesizer(nil)
	require.True(t, s.IsException("'ve"))
	require.True(t, s.IsException("n't"))
	require.True(t, s.IsException("ne'er"))
	require.False(t, s.IsException("was"))
}

func TestEnglishSynthesizer_RemoveExceptions(t *testing.T) {
	s := NewEnglishSynthesizer(nil)
	require.Equal(t, []string{"was"}, s.removeExceptions([]string{"was", "'ve", "n't"}))
}

func strp(s string) *string { return &s }
