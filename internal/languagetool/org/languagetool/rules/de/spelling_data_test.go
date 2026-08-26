package de

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSpellingData_getCoherencyMapBody ports SpellingData map construction (body mode).
func TestSpellingData_getCoherencyMapBody(t *testing.T) {
	content := `# c
Schluß;Schluss
alt;neu
`
	d, err := LoadSpellingDataBoth(content, "test.csv", nil)
	require.NoError(t, err)
	v, ok := d.Lookup("Schluß")
	require.True(t, ok)
	require.Equal(t, "Schluss", v)
	require.Equal(t, "neu", d.Map()["alt"])
	// sentence start capitalizes lowercase pairs
	require.Equal(t, "Neu", d.SentenceStartMap()["Alt"])
	// body mode keeps lowercase keys as-is (not title-cased)
	require.Equal(t, "neu", d.Map()["alt"])
	_, hasAltTitle := d.Map()["Alt"]
	require.False(t, hasAltTitle)

	_, err = LoadSpellingDataBoth("same;same\n", "x", nil)
	require.Error(t, err)
}

// TestSpellingData_ExpandForms ports SpellingData ß→ss expansion via synthesizeForPosTags (Java L68–75).
func TestSpellingData_ExpandForms(t *testing.T) {
	content := "Rußland;Russland\n"
	expand := func(old string) []string {
		if old == "Rußland" {
			// Schlüsse has "ss" — Java skips (new-spelling noise from old lemma Schluß).
			return []string{"Rußlands", "Schlüsse"}
		}
		return nil
	}
	d, err := LoadSpellingDataBoth(content, "t.csv", expand)
	require.NoError(t, err)
	require.Equal(t, "Russland", d.Map()["Rußland"])
	require.Equal(t, "Russlands", d.Map()["Rußlands"])
	_, ok := d.Map()["Schlüsse"]
	require.False(t, ok)
}

// TestSpellingData_SanityChecks ports SpellingData.sanityChecks.
func TestSpellingData_SanityChecks(t *testing.T) {
	_, err := LoadSpellingDataBoth("a;b\nb;a\n", "x", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "contradictory")

	_, err = LoadSpellingDataBoth("a;b\na;c\n", "x", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate")

	_, err = LoadSpellingDataBoth("onlyone\n", "x", nil)
	require.Error(t, err)
	require.Contains(t, strings.ToLower(err.Error()), "unexpected format")
}

// TestSpellingData_OfficialAltNeu loads Java FILE_PATH "/de/alt_neu.csv" with real GermanSynthesizer.
func TestSpellingData_OfficialAltNeu(t *testing.T) {
	require.NotNil(t, openDiscoveredGermanSynthesizer(),
		"german_synth.dict required for SpellingData ß→ss expansion (Java GermanSynthesizer.INSTANCE)")

	sd := NewSpellingData("/de/alt_neu.csv")
	require.NotNil(t, sd.GetTrie())
	require.NotNil(t, sd.GetSentenceStartTrie())

	// Direct CSV pairs (body map)
	require.Equal(t, "Schluss", sd.Map()["Schluß"])
	require.Equal(t, "Kuss", sd.Map()["Kuß"])
	require.Equal(t, "Russland", sd.Map()["Rußland"])
	require.Equal(t, "Foto", sd.Map()["Photo"])
	// multi-suggestion value
	require.Equal(t, "schwarzweiß malen|schwarz-weiß malen", sd.Map()["schwarzweißmalen"])

	// ß→ss expansion via GermanSynthesizer (Java SpellingData L68–75)
	require.Equal(t, "Russlands", sd.Map()["Rußlands"],
		"SpellingData must expand Rußland→Rußlands via synthesizer like Java")
	// Schlüsse must not be mapped from Schluß expansion (contains "ss")
	_, schluess := sd.Map()["Schlüsse"]
	require.False(t, schluess)

	// Sentence-start: lowercase pair läßt;lässt → Läßt;Lässt
	require.Equal(t, "Lässt", sd.SentenceStartMap()["Läßt"])
	// body still has lowercase
	require.Equal(t, "lässt", sd.Map()["läßt"])
}

// TestSpellingData_ParseText ports AhoCorasickDoubleArrayTrie.parseText hit offsets (UTF-16).
func TestSpellingData_ParseText(t *testing.T) {
	content := "Schluß;Schluss\nPhoto;Foto\n"
	d, err := LoadSpellingDataBoth(content, "t.csv", nil)
	require.NoError(t, err)

	hits := d.GetTrie().ParseText("Ein Schluß und Photo.")
	require.NotEmpty(t, hits)
	// Find Schluß hit
	var foundSchluss, foundPhoto bool
	for _, h := range hits {
		if h.Value == "Schluss" {
			foundSchluss = true
			covered := substringByUTF16("Ein Schluß und Photo.", h.Begin, h.End)
			require.Equal(t, "Schluß", covered)
		}
		if h.Value == "Foto" {
			foundPhoto = true
			covered := substringByUTF16("Ein Schluß und Photo.", h.Begin, h.End)
			require.Equal(t, "Photo", covered)
		}
	}
	require.True(t, foundSchluss)
	require.True(t, foundPhoto)

	// No substring match for "Photo" inside "Photons" — OldSpelling ignores via boundary,
	// but trie still reports the hit (Java same; ignoreMatch filters).
	hits2 := d.GetTrie().ParseText("Photons")
	require.NotEmpty(t, hits2)
	require.Equal(t, "Foto", hits2[0].Value)
	require.Equal(t, 0, hits2[0].Begin)
	require.Equal(t, 5, hits2[0].End) // "Photo" is 5 UTF-16 units
}

// TestSpellingData_SentenceStartTrieParse ports sentence-start trie only matching title case.
func TestSpellingData_SentenceStartTrieParse(t *testing.T) {
	content := "läßt;lässt\n"
	d, err := LoadSpellingDataBoth(content, "t.csv", nil)
	require.NoError(t, err)
	// body: lowercase key
	require.Equal(t, "lässt", d.Map()["läßt"])
	// sentence start: Läßt → Lässt
	require.Equal(t, "Lässt", d.SentenceStartMap()["Läßt"])

	hits := d.GetSentenceStartTrie().ParseText("Läßt du das bitte")
	require.NotEmpty(t, hits)
	require.Equal(t, 0, hits[0].Begin)
	require.Equal(t, "Lässt", hits[0].Value)

	// body trie finds lowercase only
	bodyHits := d.GetTrie().ParseText("läßt du")
	require.NotEmpty(t, bodyHits)
	require.Equal(t, "lässt", bodyHits[0].Value)
}

// TestSpellingData_NoExpandWhenNotSsReplacement: ß pairs that are not pure ß→ss skip expand.
func TestSpellingData_NoExpandWhenNotSsReplacement(t *testing.T) {
	// e.g. if new spelling is not old with ß→ss, expand block is skipped
	content := "Maß;Massstab\n" // not Maß→Mass (ss replacement of whole string)
	called := false
	expand := func(old string) []string {
		called = true
		return []string{"Maßes"}
	}
	d, err := LoadSpellingDataBoth(content, "t.csv", expand)
	require.NoError(t, err)
	require.False(t, called, "expand only when old.with ß→ss equals new")
	require.Equal(t, "Massstab", d.Map()["Maß"])
	_, ok := d.Map()["Maßes"]
	require.False(t, ok)
}

// TestLoadSpellingDataBoth keeps the historical test name for discoverability.
func TestLoadSpellingDataBoth(t *testing.T) {
	TestSpellingData_getCoherencyMapBody(t)
}

// TestLoadSpellingDataBoth_ExpandForms keeps the historical test name.
func TestLoadSpellingDataBoth_ExpandForms(t *testing.T) {
	TestSpellingData_ExpandForms(t)
}
