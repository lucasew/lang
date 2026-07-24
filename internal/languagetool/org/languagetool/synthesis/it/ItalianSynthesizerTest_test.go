package it

// Twin of ItalianSynthesizerTest — full IT synth dict deferred; manual synth smoke.
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

func TestItalianSynthesizer_NoTests(t *testing.T) {
	// Manual synth format: form\tlemma\tpos
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader("sono\tessere\tVER:ind+pres+1+s\n"))
	require.NoError(t, err)
	s := NewItalianSynthesizer(manual)
	require.Equal(t, "/it/italian_synth.dict", s.ResourceFileName)
	lemma := "essere"
	tag := "VER:ind+pres+1+s"
	tok := languagetool.NewAnalyzedToken("essere", &tag, &lemma)
	forms, err := s.Synthesize(tok, tag)
	require.NoError(t, err)
	require.Contains(t, forms, "sono")
}
