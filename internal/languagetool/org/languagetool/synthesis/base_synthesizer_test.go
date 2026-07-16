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

func TestJLanguageToolConstants(t *testing.T) {
	// compile-time presence via languagetool package constants tested elsewhere
}
