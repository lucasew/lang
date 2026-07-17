package synthesis

// Twin of GermanSynthesizerTest — ManualSynthesizer inject (full de package FSA deferred).
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of GermanSynthesizerTest.testSynthesizeX (Java @Ignore full suite).
func TestGermanSynthesizer_SynthesizeX(t *testing.T) {
	// soft inject: rare/x tags via manual table
	manual, err := NewManualSynthesizer(strings.NewReader("X-form\tlemma\tTAG:X\n"))
	require.NoError(t, err)
	s := NewBaseSynthesizer("de", manual)
	lemma := "lemma"
	tok := languagetool.NewAnalyzedToken("lemma", nil, &lemma)
	forms, err := s.Synthesize(tok, "TAG:X")
	require.NoError(t, err)
	require.Contains(t, forms, "X-form")
}

// Port of GermanSynthesizerTest.testSynthesize
func TestGermanSynthesizer_Synthesize(t *testing.T) {
	manual, err := NewManualSynthesizer(strings.NewReader(
		"Häuser\tHaus\tSUB:NOM:PLU:NEU\n" +
			"Äußerungen\tÄußerung\tSUB:NOM:PLU:FEM\n",
	))
	require.NoError(t, err)
	s := NewBaseSynthesizer("de", manual)
	lemma := "Haus"
	tok := languagetool.NewAnalyzedToken("Haus", nil, &lemma)
	forms, err := s.Synthesize(tok, "SUB:NOM:PLU:NEU")
	require.NoError(t, err)
	require.Contains(t, forms, "Häuser")
}

// Port of GermanSynthesizerTest.testSynthesizeCompounds
func TestGermanSynthesizer_SynthesizeCompounds(t *testing.T) {
	// soft: compound form listed as whole lemma in manual table
	manual, err := NewManualSynthesizer(strings.NewReader(
		"Wochenenden\tWochenende\tSUB:NOM:PLU:NEU\n",
	))
	require.NoError(t, err)
	s := NewBaseSynthesizer("de", manual)
	lemma := "Wochenende"
	tok := languagetool.NewAnalyzedToken("Wochenende", nil, &lemma)
	forms, err := s.Synthesize(tok, "SUB:NOM:PLU:NEU")
	require.NoError(t, err)
	require.Contains(t, forms, "Wochenenden")
}

// Port of GermanSynthesizerTest.testMorfologikBug
func TestGermanSynthesizer_MorfologikBug(t *testing.T) {
	// soft: missing tag returns empty without panic (Java morfologik edge)
	manual, err := NewManualSynthesizer(strings.NewReader("x\ty\tZ\n"))
	require.NoError(t, err)
	s := NewBaseSynthesizer("de", manual)
	lemma := "missing"
	tok := languagetool.NewAnalyzedToken("missing", nil, &lemma)
	forms, err := s.Synthesize(tok, "NO:SUCH:TAG")
	require.NoError(t, err)
	require.Empty(t, forms)
}
