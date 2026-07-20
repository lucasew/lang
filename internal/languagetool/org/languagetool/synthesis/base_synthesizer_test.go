package synthesis

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestBaseSynthesizer(t *testing.T) {
	man, err := NewManualSynthesizer(strings.NewReader("dogs\tdog\tNNS\n"))
	require.NoError(t, err)
	s := NewBaseSynthesizer("en", man)
	lemma := "dog"
	tok := languagetool.NewAnalyzedToken("dog", nil, &lemma)
	forms, err := s.Synthesize(tok, "NNS")
	require.NoError(t, err)
	require.Equal(t, []string{"dogs"}, forms)
}

// Port of Java BaseSynthesizer.synthesizeForPosTags (accept-all and filter).
func TestBaseSynthesizer_SynthesizeForPosTags(t *testing.T) {
	man, err := NewManualSynthesizer(strings.NewReader(
		"Rußlands\tRußland\tSUB:GEN:SIN:NEU\n" +
			"Rußland\tRußland\tSUB:NOM:SIN:NEU\n" +
			"dogs\tdog\tNNS\n",
	))
	require.NoError(t, err)
	s := NewBaseSynthesizer("de", man)

	all := s.SynthesizeForPosTags("Rußland", func(string) bool { return true })
	require.ElementsMatch(t, []string{"Rußlands", "Rußland"}, all)

	gen := s.SynthesizeForPosTags("Rußland", func(tag string) bool {
		return strings.Contains(tag, "GEN")
	})
	require.Equal(t, []string{"Rußlands"}, gen)

	require.Empty(t, s.SynthesizeForPosTags("missing", func(string) bool { return true }))
	require.Empty(t, s.SynthesizeForPosTags("Rußland", nil))
}

func TestJLanguageToolConstants(t *testing.T) {
	// compile-time presence via languagetool package constants tested elsewhere
}

func TestBaseSynthesizer_RemoveExceptions_IsExceptionFn(t *testing.T) {
	man, err := NewManualSynthesizer(strings.NewReader(
		"was\tbe\tVBD\n" +
			"'ve\tbe\tVBD\n" +
			"n't\tnot\tRB\n",
	))
	require.NoError(t, err)
	s := NewBaseSynthesizer("en", man)
	s.IsExceptionFn = func(w string) bool {
		return strings.HasPrefix(w, "'") || w == "n't"
	}
	lemma := "be"
	tok := languagetool.NewAnalyzedToken("be", nil, &lemma)
	got, err := s.Synthesize(tok, "VBD")
	require.NoError(t, err)
	require.Equal(t, []string{"was"}, got) // 've filtered
	// SynthesizeForPosTags also filters
	all := s.SynthesizeForPosTags("be", func(string) bool { return true })
	require.Equal(t, []string{"was"}, all)
}
