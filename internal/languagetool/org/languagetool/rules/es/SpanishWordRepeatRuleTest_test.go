package es

// Twin of languagetool-language-modules/es/src/test/java/org/languagetool/rules/es/SpanishWordRepeatRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSpanishWordRepeatRule_Rule(t *testing.T) {
	rule := NewSpanishWordRepeatRule(map[string]string{"repetition": "Repetición"})
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Bienvenido/a a LanguageTool."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("HUCHA-GANGA.ES es la web de referencia."))))
	// real error still fires
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Esto es es un error."))))
}
