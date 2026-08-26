package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenFrenchSynthesizer_Missing(t *testing.T) {
	require.Nil(t, OpenFrenchSynthesizerFromDir(""))
	require.Nil(t, OpenFrenchSynthesizerFromDictPath(""))
	require.Nil(t, OpenFrenchSynthesizerFromDir(t.TempDir()))
}

// Open path fail-closed + successful load from discovered dict.
func TestOpenFrenchSynthesizer_FromDict(t *testing.T) {
	p := DiscoverFrenchSynthDict()
	if p == "" {
		t.Skip("french_synth.dict not discoverable")
	}
	s := OpenFrenchSynthesizerFromDictPath(p)
	require.NotNil(t, s)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Equal(t, FrenchSynthResource, s.ResourceFileName)
	require.Equal(t, FrenchTagsFile, s.TagFileName)
	require.Equal(t, FrenchSorFile, s.SorFileName)
	// Inspiration supplements when manuals are not beside third_party dict.
	require.NotNil(t, s.Manual, "added.txt from inspiration module")
	require.NotNil(t, s.IsExceptionFn)

	lemma := "nager"
	tok := languagetool.NewAnalyzedToken("nager", strp("V"), &lemma)
	forms, err := s.Synthesize(tok, "V ind pres 3 p")
	require.NoError(t, err)
	require.Equal(t, []string{"nagent"}, forms)

	// unknown → empty (not invent)
	forms, err = s.Synthesize(tok, "ZZZZ")
	require.NoError(t, err)
	require.Empty(t, forms)
}

// Instance singleton binds INSTANCE to loaded official synthesizer.
func TestFrenchSynthesizer_Instance(t *testing.T) {
	s := Instance()
	if s == nil {
		t.Skip("french_synth.dict / official FR synth resources not available")
	}
	require.Same(t, s, Instance())
	require.Same(t, s, INSTANCE)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
}

func TestFrenchSynthesizer_IsException(t *testing.T) {
	s := NewFrenchSynthesizer(nil)
	require.Equal(t, FrenchSynthResource, s.ResourceFileName)
	require.Equal(t, FrenchTagsFile, s.TagFileName)
	require.Equal(t, FrenchSorFile, s.SorFileName)
	// Java isException
	require.True(t, s.IsException("qqchose"))
	require.True(t, s.IsException("informè")) // trailing è not in allowlist
	require.False(t, s.IsException("burkinabè"))
	require.False(t, s.IsException("koinè"))
	require.False(t, s.IsException("épistémè"))
	require.False(t, s.IsException("maison"))
}

func strp(s string) *string { return &s }
