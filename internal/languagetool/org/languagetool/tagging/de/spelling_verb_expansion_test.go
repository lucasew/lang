package de

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestSpellingVerbExpansion_ZuAndNominalized(t *testing.T) {
	ex, err := LoadSpellingVerbExpansion(strings.NewReader("weg_delegieren\n"))
	require.NoError(t, err)
	vi, ok := ex.LookupVerb("wegdelegieren")
	require.True(t, ok)
	require.Equal(t, "weg", vi.Prefix)
	require.Equal(t, "delegieren", vi.VerbBaseform)

	vi2, ok := ex.LookupVerb("wegzudelegieren")
	require.True(t, ok)
	require.Equal(t, "zu", vi2.Infix)

	// fixed zu reading
	tags := ex.Tag("wegzudelegieren")
	require.NotEmpty(t, tags)
	require.Equal(t, "VER:EIZ:NON", tags[0].PosTag)

	// nominalized
	nom := ex.Tag("Wegdelegieren")
	require.Len(t, nom, 3)
	require.Equal(t, "SUB:NOM:SIN:NEU:INF", nom[0].PosTag)
	gen := ex.Tag("Wegdelegierens")
	require.NotEmpty(t, gen)
	require.Equal(t, "SUB:GEN:SIN:NEU:INF", gen[0].PosTag)
}

func TestGermanTagger_VerbAndWeise(t *testing.T) {
	// base tagger knows "ideal" as ADJ for weise
	wt := tagging.MapWordTagger{
		"ideal": {tagging.NewTaggedWord("ideal", "ADJ:PRD:GRU")},
		"geben": {tagging.NewTaggedWord("geben", "VER:INF:NON")},
	}
	tagger := NewGermanTagger(wt)
	vex, err := LoadSpellingVerbExpansion(strings.NewReader("herum_geben\n"))
	require.NoError(t, err)
	tagger.SetSpellingVerbExpansion(vex)

	// prefix + infinitive
	rd := tagger.Lookup("herumgeben")
	require.NotNil(t, rd)
	// should pick VER from "geben" with lemma herumgeben
	found := false
	for _, r := range rd.GetReadings() {
		if r.GetPOSTag() != nil && strings.HasPrefix(*r.GetPOSTag(), "VER:") {
			found = true
			require.NotNil(t, r.GetLemma())
			require.Equal(t, "herumgeben", *r.GetLemma())
		}
	}
	require.True(t, found, "expected VER reading for herumgeben")

	// zu form
	rd2 := tagger.Lookup("herumzugeben")
	require.NotNil(t, rd2)
	require.NotNil(t, rd2.GetReadings()[0].GetPOSTag())
	require.Equal(t, "VER:EIZ:NON", *rd2.GetReadings()[0].GetPOSTag())

	// nominalized
	rd3 := tagger.Lookup("Herumgeben")
	require.NotNil(t, rd3)
	require.NotNil(t, rd3.GetReadings()[0].GetPOSTag())
	require.Contains(t, *rd3.GetReadings()[0].GetPOSTag(), "SUB:")

	// idealerweise
	rd4 := tagger.Lookup("idealerweise")
	require.NotNil(t, rd4)
	require.NotNil(t, rd4.GetReadings()[0].GetPOSTag())
	require.True(t, strings.HasPrefix(*rd4.GetReadings()[0].GetPOSTag(), "ADJ:"))
}

func TestSpellingVerbExpansion_FromOfficialFile(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	path := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/hunspell/spelling.txt")
	ex, err := LoadSpellingVerbExpansionFromFile(path)
	if err != nil {
		t.Skipf("spelling.txt: %v", err)
	}
	require.Greater(t, ex.Size(), 100)
	_, ok := ex.LookupVerb("wegdelegieren")
	require.True(t, ok)
}
