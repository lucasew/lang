package commandline

import (
	"os"
	"path/filepath"
	"testing"

	plsynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/pl"
	"github.com/stretchr/testify/require"
)

// Twin of Java *Synthesizer.RESOURCE_FILENAME basenames used by createDefaultSynthesizer.
func TestLanguageSynthDictNames_JavaResourceBasenames(t *testing.T) {
	// Every Java createDefaultSynthesizer resource basename must be listed.
	want := map[string]string{
		"en":  "english_synth.dict",
		"pl":  "polish_synth.dict",
		"de":  "german_synth.dict",
		"fr":  "french_synth.dict",
		"nl":  "dutch_synth.dict",
		"pt":  "portuguese_synth.dict",
		"es":  "es-ES_synth.dict",
		"ca":  "ca-ES_synth.dict",
		"it":  "italian_synth.dict",
		"ru":  "russian_synth.dict",
		"ro":  "romanian_synth.dict",
		"sk":  "slovak_synth.dict",
		"sv":  "swedish_synth.dict",
		"el":  "greek_synth.dict",
		"gl":  "galician_synth.dict",
		"ar":  "arabic_synth.dict",
		"uk":  "ukrainian_synth.dict",
		"ga":  "irish_synth.dict",
		"crh": "crimean_tatar_synth.dict",
		"sr":  "serbian_synth.dict",
	}
	for code, name := range want {
		require.Equal(t, name, languageSynthDictNames[code], "lang %s", code)
	}
	// No invent soft basenames.
	require.NotContains(t, languageSynthDictNames, "soft")
}

func TestDiscoverLanguageSynthDict_PolishInspiration(t *testing.T) {
	p := DiscoverLanguageSynthDict(nil, "pl")
	if p == "" {
		t.Skip("polish_synth.dict not found under inspiration")
	}
	require.True(t, filepath.IsAbs(p) || p != "")
	st, err := os.Stat(p)
	require.NoError(t, err)
	require.True(t, st.Mode().IsRegular())
	require.Equal(t, "polish_synth.dict", filepath.Base(p))
}

func TestOpenLanguageSynthesizer_PolishGetPosTagCorrection(t *testing.T) {
	p := DiscoverLanguageSynthDict(nil, "pl")
	if p == "" {
		t.Skip("polish_synth.dict not found")
	}
	s := OpenLanguageSynthesizer("pl", p)
	require.NotNil(t, s)
	// Must be PolishSynthesizer so setpos getPosTagCorrection expands a.z segments.
	ps, ok := s.(*plsynth.PolishSynthesizer)
	require.True(t, ok, "createDefaultSynthesizer(pl) must be PolishSynthesizer")
	got := ps.GetPosTagCorrection("adj:a.b:sg")
	require.Contains(t, got, ".*|.*", "Polish getPosTagCorrection must expand dotted segments")
}

func TestOpenLanguageSynthesizer_Missing(t *testing.T) {
	require.Nil(t, OpenLanguageSynthesizer("pl", ""))
	require.Nil(t, OpenLanguageSynthesizer("pl", filepath.Join(t.TempDir(), "nope.dict")))
}

func TestLanguageSynthInspirationRels_Serbian(t *testing.T) {
	rels := languageSynthInspirationRels("sr", "serbian_synth.dict")
	require.GreaterOrEqual(t, len(rels), 2)
	require.Contains(t, rels[0], "ekavian")
	require.Contains(t, rels[1], "jekavian")
}
