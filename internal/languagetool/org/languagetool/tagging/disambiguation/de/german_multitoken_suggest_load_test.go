package de

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenGermanMultitokenSuggest_SettingsAndIgnore(t *testing.T) {
	// Tiny in-memory stand-in of official format (not invent for engine input —
	// proves OpenGermanMultitokenSuggest applies Java constructor settings).
	r := strings.NewReader("Foo Bar/S\n")
	c, err := OpenGermanMultitokenSuggest(r)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.True(t, c.AddIgnoreSpelling)
	require.Contains(t, c.Lines, "Foo Bar")
	require.Contains(t, c.Lines, "Foo Bars") // /S expansion
}

func TestGermanMultitokenSuggest_ProcessCachedOfficial(t *testing.T) {
	if DiscoverGermanMultitokenSuggest() == "" {
		t.Skip("official multitoken-suggest.txt not discoverable")
	}
	a := GermanMultitokenSuggest()
	b := GermanMultitokenSuggest()
	require.NotNil(t, a)
	require.Same(t, a, b, "process-cached singleton")
	// Official MultitokenSuggest entries
	require.Contains(t, a.Lines, "New York")
	require.Contains(t, a.Lines, "New Yorks") // /S expansion
	require.Contains(t, a.Lines, "à la carte")
	require.Contains(t, a.Lines, "Alma Mater")
	require.Contains(t, a.Lines, "Deus ex Machina")
	require.Contains(t, a.Lines, "Osama bin Laden")
	require.Contains(t, a.Lines, "Osama bin Ladens") // /S
	require.Contains(t, a.Lines, "Human-centered Design")
	// Wrong suffix forms must not appear from /S-only lines
	require.NotContains(t, a.Lines, "New Yorkn")
	require.NotContains(t, a.Lines, "Osama bin Ladenn")

	// Wired on NewGermanRuleDisambiguator
	d := NewGermanRuleDisambiguator()
	require.NotNil(t, d.MultitokenSuggest)
	require.Same(t, a, d.MultitokenSuggest)
}

func TestGermanMultitokenSuggest_DisambiguateOfficialPhrases(t *testing.T) {
	if DiscoverGermanMultitokenSuggest() == "" {
		t.Skip("official multitoken-suggest.txt not discoverable")
	}
	// Isolate MultitokenSuggest stage (do not re-claim MultitokenIgnore/Global).
	s := GermanMultitokenSuggest()
	require.NotNil(t, s)
	d := &GermanRuleDisambiguator{MultitokenSuggest: s}

	cases := []struct {
		parts []string
		want  bool
		label string
	}{
		{[]string{"New", "York"}, true, "New York"},
		{[]string{"New", "Yorks"}, true, "New Yorks /S"},
		{[]string{"à", "la", "carte"}, true, "à la carte"},
		{[]string{"Alma", "Mater"}, true, "Alma Mater"},
		{[]string{"Deus", "ex", "Machina"}, true, "Deus ex Machina"},
		{[]string{"Osama", "bin", "Laden"}, true, "Osama bin Laden"},
		{[]string{"Osama", "bin", "Ladens"}, true, "Osama bin Ladens /S"},
		{[]string{"Human-centered", "Design"}, true, "Human-centered Design"},
		// Negatives: non-listed multi-token random
		{[]string{"Zxqwv", "Plmnb"}, false, "random non-listed"},
		// Wrong suffix: /S expands s only, not n
		{[]string{"New", "Yorkn"}, false, "New Yorkn wrong suffix"},
		// allowFirstCapitalized=true: first-cap of lowercase official entry
		{[]string{"A", "cappella"}, true, "A cappella first-cap allowed"},
		// exact lowercase official entry
		{[]string{"a", "cappella"}, true, "a cappella exact"},
		// allowAllUppercase=true
		{[]string{"NEW", "YORK"}, true, "NEW YORK all-upper"},
	}
	for _, tc := range cases {
		out := d.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens(tc.parts...)))
		requireAllContentIgnored(t, out, tc.want, tc.label)
	}

	// NewGermanRuleDisambiguator wires MultitokenSuggest and still ignores official phrases.
	full := NewGermanRuleDisambiguator()
	require.NotNil(t, full.MultitokenSuggest)
	out := full.Disambiguate(languagetool.NewAnalyzedSentence(multiwordTokens("New", "York")))
	requireAllContentIgnored(t, out, true, "wired NewGermanRuleDisambiguator New York")
}
