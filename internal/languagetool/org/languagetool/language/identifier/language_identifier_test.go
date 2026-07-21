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

// Twin of LanguageIdentifier.cleanAndShortenText: maxLength is UTF-16 units.
func TestCleanAndShortenText_UTF16MaxLength(t *testing.T) {
	// maxLength 3 on "café" (4 UTF-16 units) → substring(0,3) = "caf"
	got := BaseLanguageIdentifier{MaxLength: 3}.CleanAndShortenText("café")
	require.Equal(t, "caf", got, "Java text.substring(0, maxLength) is UTF-16")
	// full "café" kept when maxLength >= 4
	require.Equal(t, "café", BaseLanguageIdentifier{MaxLength: 4}.CleanAndShortenText("café"))
	// emoji is 2 UTF-16 units; maxLength 1 keeps first unit only
	e := BaseLanguageIdentifier{MaxLength: 1}.CleanAndShortenText("😀x")
	require.Equal(t, 1, javaStringLen(e))
}
