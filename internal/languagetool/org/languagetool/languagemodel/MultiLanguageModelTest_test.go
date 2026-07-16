package languagemodel

// Twin of MultiLanguageModelTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	"github.com/stretchr/testify/require"
)

type fakeLM struct{ v float64 }

func (f fakeLM) GetPseudoProbability([]string) ngrams.Probability {
	return ngrams.NewProbabilitySimple(f.v, 0.5)
}
func (f fakeLM) Close() error { return nil }

func TestMultiLanguageModel_Test(t *testing.T) {
	lm := NewMultiLanguageModel([]LanguageModel{fakeLM{0.5}, fakeLM{0.2}})
	defer lm.Close()
	p := lm.GetPseudoProbability([]string{"foo", "bar", "blah"})
	require.InDelta(t, 0.7, p.GetProb(), 0.01)
	require.InDelta(t, 0.5, float64(p.GetCoverage()), 0.01)
}
