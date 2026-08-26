package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Open path fail-closed + successful load from discovered dict.
func TestOpenEnglishSynthesizer_FromDict(t *testing.T) {
	require.Nil(t, OpenEnglishSynthesizerFromDir(""))
	require.Nil(t, OpenEnglishSynthesizerFromDir(t.TempDir()))
	require.Nil(t, OpenEnglishSynthesizerFromDictPath(""))

	p := DiscoverEnglishSynthDict()
	if p == "" {
		t.Skip("english_synth.dict not discoverable")
	}
	s := OpenEnglishSynthesizerFromDictPath(p)
	require.NotNil(t, s)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Equal(t, EnglishSynthResource, s.ResourceFileName)
	require.Equal(t, EnglishTagsFile, s.TagFileName)
	require.Equal(t, EnglishSorFile, s.SorFileName)
	// Inspiration supplements when manuals are not beside third_party dict.
	require.NotNil(t, s.Manual, "added.txt from inspiration module")
	require.NotNil(t, s.NumberSpeller, "en.sor number speller")

	lemma := "president"
	tok := languagetool.NewAnalyzedToken("president", strp("NN"), &lemma)
	forms, err := s.Synthesize(tok, "NNS")
	require.NoError(t, err)
	require.Equal(t, []string{"presidents"}, forms)

	lemma = "be"
	tok = languagetool.NewAnalyzedToken("be", strp("VB"), &lemma)
	forms, err = s.Synthesize(tok, "VBD")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"was", "were"}, forms)

	// unknown → empty (not invent)
	forms, err = s.Synthesize(tok, "ZZZZ")
	require.NoError(t, err)
	require.Empty(t, forms)
}

// Instance singleton binds INSTANCE to loaded official synthesizer.
func TestEnglishSynthesizer_Instance(t *testing.T) {
	s := Instance()
	if s == nil {
		t.Skip("english_synth.dict / official EN synth resources not available")
	}
	require.Same(t, s, Instance())
	require.Same(t, s, INSTANCE)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
}

func strp(s string) *string { return &s }
