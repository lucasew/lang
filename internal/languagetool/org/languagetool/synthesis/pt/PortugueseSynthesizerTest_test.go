package pt

// Twin of PortugueseSynthesizerTest — full dict deferred; ManualSynthesizer path.
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

func TestPortugueseSynthesizer_NoTests(t *testing.T) {
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader("forms\tlemma\tTAG\n"))
	require.NoError(t, err)
	s := NewPortugueseSynthesizer(manual)
	require.Equal(t, PortugueseSynthDict, s.ResourceFileName)
	lemma, tag := "lemma", "TAG"
	tok := languagetool.NewAnalyzedToken("lemma", &tag, &lemma)
	got, err := s.Synthesize(tok, tag)
	require.NoError(t, err)
	require.Contains(t, got, "forms")
}
