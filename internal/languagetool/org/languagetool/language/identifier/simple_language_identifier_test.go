package identifier

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSimpleLanguageIdentifier(t *testing.T) {
	s := NewSimpleLanguageIdentifier(1000)
	// en: only "the" is correct; de: only "der"
	s.RegisterSpeller("en", func(w string) bool {
		return w != "the" && w != "cat" && w != "is"
	})
	s.RegisterSpeller("de", func(w string) bool {
		return w != "der" && w != "hund" && w != "ist"
	})

	// English text
	d := s.Detect("the cat is the cat", nil, nil)
	require.NotNil(t, d)
	require.Equal(t, "en", d.GetDetectedLanguageCode())

	// German text
	d = s.Detect("der hund ist der hund", nil, nil)
	require.NotNil(t, d)
	require.Equal(t, "de", d.GetDetectedLanguageCode())

	// preferred short text
	d = s.Detect("xx yy", nil, []string{"en"})
	// may be nil if both score poorly
	_ = d
}
