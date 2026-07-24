package en

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func findEnglishSynthDict(t *testing.T) string {
	t.Helper()
	dir, _ := filepath.Abs(".")
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, "third_party/english-pos-dict/org/languagetool/resource/en/english_synth.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
		dir = filepath.Dir(dir)
	}
	return ""
}

func TestOpenEnglishSynthesizerFromDict_RealDict(t *testing.T) {
	p := findEnglishSynthDict(t)
	if p == "" {
		t.Skip("english_synth.dict not in tree")
	}
	s := OpenEnglishSynthesizerFromDictPath(p)
	require.NotNil(t, s)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags, "english_tags.txt should load")

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
