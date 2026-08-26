package synthesis

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestFuncSynthesizer(t *testing.T) {
	s := FuncSynthesizer{
		Synth: func(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
			return []string{token.GetToken() + "s"}, nil
		},
	}
	pos := "NN"
	tok := languagetool.NewAnalyzedToken("cat", &pos, nil)
	forms, err := s.Synthesize(tok, "NNS")
	require.NoError(t, err)
	require.Equal(t, []string{"cats"}, forms)
}
