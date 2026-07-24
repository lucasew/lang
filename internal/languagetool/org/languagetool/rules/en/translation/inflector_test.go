package translation

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

func TestInflector(t *testing.T) {
	synth := synthesis.FuncSynthesizer{
		Synth: func(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
			return []string{token.GetToken() + "s"}, nil
		},
		SynthRE: func(token *languagetool.AnalyzedToken, posTag string, _ bool) ([]string, error) {
			return []string{token.GetToken() + "s"}, nil
		},
	}
	inf := NewInflector(synth)
	got := inf.Inflect("pump", "SUB:NOM:PLU:NEU")
	require.Equal(t, []string{"pumps"}, got)

	got2 := inf.Inflect("tire pump", "SUB:NOM:PLU:NEU")
	require.Equal(t, []string{"tire pumps"}, got2)

	got3 := inf.Inflect("run", "")
	require.Equal(t, []string{"run"}, got3)
}
