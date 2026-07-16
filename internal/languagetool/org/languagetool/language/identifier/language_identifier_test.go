package identifier

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCleanAndShortenText(t *testing.T) {
	b := NewBaseLanguageIdentifier(100)
	out := b.CleanAndShortenText("Hello https://example.com world @user\n-- \nsig")
	require.NotContains(t, out, "https://")
	require.NotContains(t, out, "@user")
	require.Contains(t, out, "Hello")
}

func TestSimpleSpellScoreIdentifier(t *testing.T) {
	id := NewSimpleSpellScoreIdentifier(map[string]func(string) bool{
		"en": func(w string) bool { return w == "the" || w == "cat" },
		"de": func(w string) bool { return w == "der" || w == "Hund" },
	})
	d := id.Detect("the cat", nil, nil)
	require.NotNil(t, d)
	require.Equal(t, "en", d.GetDetectedLanguageCode())
}

func TestLanguageIdentifierService(t *testing.T) {
	Instance.Clear()
	require.Nil(t, Instance.GetInitialized())
	id := NewMapLanguageIdentifier(100, func(text string, pref []string) map[string]float64 {
		return map[string]float64{"fr": 0.9}
	})
	Instance.SetSimple(id)
	require.Equal(t, "fr", Instance.GetInitialized().Detect("bonjour", nil, nil).GetDetectedLanguageCode())
	Instance.Clear()
	require.True(t, CanLanguageBeDetected("en", []string{"en", "de"}, nil))
}
