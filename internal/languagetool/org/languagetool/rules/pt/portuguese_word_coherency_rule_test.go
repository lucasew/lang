package pt

// Example-based twin for PortugueseWordCoherencyRule (no dedicated Java unit test).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPortugueseWordCoherencyRule(t *testing.T) {
	rule := NewPortugueseWordCoherencyRule(nil)
	// Java example: duradouro vs duradoiro
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Foi um período duradouro e marcante."),
	})))
	require.Equal(t, 1, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Foi um período duradouro. Tão marcante e duradoiro dificilmente será esquecido."),
	})))
}
