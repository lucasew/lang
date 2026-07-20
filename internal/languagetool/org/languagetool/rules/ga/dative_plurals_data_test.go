package ga

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDativePluralsEntry(t *testing.T) {
	e := NewDativePluralsEntry("fearaibh", "fear", "m", "fir")
	require.Equal(t, "fir", e.GetStandard())
	e.SetEquivalent("fir")
	require.True(t, e.HasEquivalent())
	require.Equal(t, "Noun:Masc:Dat:Pl", e.GetBaseTag())
	e2 := NewDativePluralsEntry("mnáibh", "bean", "f", "mná")
	require.Equal(t, "Noun:Fem:Dat:Pl", e2.GetBaseTag())
}

func TestParseDativePluralsData(t *testing.T) {
	in := strings.NewReader("fearaibh;fear;m;fir\nbliantaibh:blianta;bliain;f;blianta\n")
	d, err := ParseDativePluralsData(in)
	require.NoError(t, err)
	require.Len(t, d.Entries, 2)
	require.Equal(t, "fir", d.GetSimpleReplacements()["fearaibh"])
	require.Equal(t, "blianta", d.GetModernisations()["bliantaibh"])
}

func TestLoadDativePluralsData(t *testing.T) {
	d := LoadDativePluralsData()
	require.NotNil(t, d)
	// embedded file should yield some entries when present
	_ = d.GetSimpleReplacements()
}

func TestDhaNoBeirtData(t *testing.T) {
	var d DhaNoBeirtData
	nums := d.GetNumberReplacements()
	require.Equal(t, "beirt", nums["dhá"])
	require.NotNil(t, d.GetDaoine())
}

func TestIrishPartialPosTagFilter(t *testing.T) {
	f := NewIrishPartialPosTagFilter(func(s string) []string { return []string{"Noun"} })
	require.NotNil(t, f)
	ClearDefaultIrishPartialPosTagger()
	f2 := NewNoDisambiguationIrishPartialPosTagFilter(nil)
	require.NotNil(t, f2)
	// fail-closed without process-wide tagger
	ok, err := f2.Accept("x", "^(x)$", "Noun", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestIrishPartialPosTagFilter_WithTagAndDisambig(t *testing.T) {
	SetDefaultIrishPartialPosTagger(func(p string) []string {
		if p == "cat" {
			return []string{"Noun"}
		}
		return nil
	})
	WireIrishFilterDisambiguator(stubGADisambig{})
	t.Cleanup(func() {
		ClearDefaultIrishPartialPosTagger()
		ClearIrishFilterDisambiguator()
	})
	f := NewIrishPartialPosTagFilter(nil)
	ok, err := f.Accept("cats", "^(cat)s$", "Noun", false, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)
}

type stubGADisambig struct{}

func (stubGADisambig) Disambiguate(s *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	return s
}
