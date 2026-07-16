package es

// Twin of languagetool-language-modules/es/src/test/java/org/languagetool/rules/es/QuestionMarkRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestQuestionMarkRule_Rule(t *testing.T) {
	rule := NewQuestionMarkRule(nil)
	assertN := func(s string, n int, sugg string) {
		t.Helper()
		matches := rule.MatchList(languagetool.AnalyzeTextLocal(s))
		require.Equal(t, n, len(matches), "text %q got %v", s, matches)
		if n >= 1 && sugg != "" {
			require.Equal(t, sugg, matches[0].GetSuggestedReplacements()[0], "text %q", s)
		}
	}

	assertN("Hola, ¿cómo estás?", 0, "")
	assertN("Hola, cómo estás?", 1, "¿cómo")
	assertN("¿Que pasa?", 0, "")
	assertN("Que pasa?", 1, "¿Que")
	assertN("Que pasa?\n", 1, "¿Que")
	assertN("¡¿Nunca tienes clases o qué?!", 0, "")
	assertN("¿Quién sabe hablar francés mejor: Tom o Mary?", 0, "")
	assertN("Esto es una prueba. Que pasa?\n\n", 1, "¿Que")
	assertN("Hola, de qué me hablas?", 1, "¿de")
	assertN("Después de todo lo que pasó, qué quieres que te diga?", 1, "¿qué")
	assertN("Pero cómo quieres que te lo diga?", 1, "¿Pero")
	assertN("Pero, cómo quieres que te lo diga?", 1, "¿cómo")
	assertN("Puedes imaginarte por qué no vino con nosotros?", 1, "¿Puedes")
	assertN("Hola, Marco: Puedes darme tu dirección de correo?", 1, "¿Puedes")

	assertN("Qué irritante!", 1, "¡Qué")
	assertN("¡Qué irritante!", 0, "")
	assertN("—Hola!", 1, "¡Hola")
	assertN("Muchas gracias!", 1, "¡Muchas")
	assertN("Tengo razón, o no?", 1, "¿o")
	assertN("Tengo razón, no?", 1, "¿no")
	assertN("Tengo razón, eh?", 1, "¿eh")
	assertN("qué me recomendarías???….", 1, "¿qué")
}
