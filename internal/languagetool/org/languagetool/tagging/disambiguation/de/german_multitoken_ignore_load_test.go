package de

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenGermanMultitokenIgnore_SettingsAndIgnore(t *testing.T) {
	// Tiny in-memory stand-in of official format (not invent for engine input —
	// proves OpenGermanMultitokenIgnore applies Java constructor settings).
	r := strings.NewReader("Foo Bar/N\n")
	c, err := OpenGermanMultitokenIgnore(r)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.True(t, c.AddIgnoreSpelling)
	require.Contains(t, c.Lines, "Foo Bar")
	require.Contains(t, c.Lines, "Foo Barn") // /N expansion
}

func TestGermanMultitokenIgnore_ProcessCachedOfficial(t *testing.T) {
	if DiscoverGermanMultitokenIgnore() == "" {
		t.Skip("official multitoken-ignore.txt not discoverable")
	}
	a := GermanMultitokenIgnore()
	b := GermanMultitokenIgnore()
	require.NotNil(t, a)
	require.Same(t, a, b, "process-cached singleton")
	// Official MultitokenIgnore entry from GermanDisambiguationTest
	require.Contains(t, a.Lines, "3-adische System")
	require.Contains(t, a.Lines, "3-adische Systeme") // /E
	require.Contains(t, a.Lines, "Kelassurier Mauer")
	require.Contains(t, a.Lines, "Kelassurier Mauern") // /N
	require.NotContains(t, a.Lines, "Kelassurier Mauers")

	// Wired on NewGermanRuleDisambiguator
	d := NewGermanRuleDisambiguator()
	require.NotNil(t, d.MultitokenIgnore)
	require.Same(t, a, d.MultitokenIgnore)
}

func TestGermanMultitokenIgnore_DisambiguateJavaCases(t *testing.T) {
	if DiscoverGermanMultitokenIgnore() == "" {
		t.Skip("official multitoken-ignore.txt not discoverable")
	}
	d := NewGermanRuleDisambiguator()
	require.NotNil(t, d.MultitokenIgnore)

	cases := []struct {
		a, b string
		want bool
	}{
		{"3-adische", "System", true},
		{"3-adische", "Systeme", true},
		{"3-adischen", "Systems", true},
		{"Kelassurier", "Mauer", true},
		{"Kelassurier", "Mauern", true},
		{"Kelassurier", "Mauers", false},
	}
	for _, tc := range cases {
		tag := languagetool.SentenceStartTagName
		toks := []*languagetool.AnalyzedTokenReadings{
			languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
			languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(tc.a, nil, nil)),
			languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
			languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(tc.b, nil, nil)),
		}
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(toks))
		got1 := out.GetTokens()[1].IsIgnoredBySpeller()
		got2 := out.GetTokens()[3].IsIgnoredBySpeller()
		require.Equal(t, tc.want, got1, "%s %s first", tc.a, tc.b)
		require.Equal(t, tc.want, got2, "%s %s second", tc.a, tc.b)
	}
}
