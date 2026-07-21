package es

// Twin of languagetool-language-modules/es/src/test/java/org/languagetool/rules/es/QuestionMarkRuleTest.java
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestQuestionMarkRule_Test(t *testing.T) {
	rule := NewQuestionMarkRule(nil)
	assertN := func(s string, n int, sugg string) {
		t.Helper()
		matches := rule.MatchList(languagetool.AnalyzeTextLocal(s))
		require.Equal(t, n, len(matches), "text %q got %v", s, formatQM(matches))
		if n >= 1 && sugg != "" {
			require.Equal(t, sugg, matches[0].GetSuggestedReplacements()[0], "text %q", s)
		}
	}
	// POS inject for FreeLing tags used by Java comma-clause reposition.
	assertNTagged := func(s string, tags map[string]string, n int, sugg string) {
		t.Helper()
		sents := languagetool.AnalyzeTextLocal(s)
		for _, sent := range sents {
			injectESTags(sent, tags)
		}
		matches := rule.MatchList(sents)
		require.Equal(t, n, len(matches), "text %q got %v", s, formatQM(matches))
		if n >= 1 && sugg != "" {
			require.Equal(t, sugg, matches[0].GetSuggestedReplacements()[0], "text %q", s)
		}
	}

	assertN("Hola, ¿cómo estás?", 0, "")
	// "cómo" after comma needs PT* for firstToken reposition
	assertNTagged("Hola, cómo estás?", map[string]string{"cómo": "PT000000"}, 1, "¿cómo")
	assertN("¿Que pasa?", 0, "")
	assertN("Que pasa?", 1, "¿Que")
	assertN("Que pasa?\n", 1, "¿Que")
	assertN("¡¿Nunca tienes clases o qué?!", 0, "")
	assertN("¿Quién sabe hablar francés mejor: Tom o Mary?", 0, "")
	assertN("Esto es una prueba. Que pasa?\n\n", 1, "¿Que")
	// de + qué → SPS00 + PT*
	assertNTagged("Hola, de qué me hablas?", map[string]string{"de": "SPS00", "qué": "PT000000"}, 1, "¿de")
	assertNTagged("Después de todo lo que pasó, qué quieres que te diga?", map[string]string{"qué": "PT000000"}, 1, "¿qué")
	// "Pero cómo" without comma: firstToken stays Pero (Java)
	assertN("Pero cómo quieres que te lo diga?", 1, "¿Pero")
	assertNTagged("Pero, cómo quieres que te lo diga?", map[string]string{"cómo": "PT000000"}, 1, "¿cómo")
	assertN("Puedes imaginarte por qué no vino con nosotros?", 1, "¿Puedes")
	// colon resets firstToken so next content word is Puedes
	assertN("Hola, Marco: Puedes darme tu dirección de correo?", 1, "¿Puedes")

	assertN("Qué irritante!", 1, "¡Qué")
	assertN("¡Qué irritante!", 0, "")
	assertN("—Hola!", 1, "¡Hola")
	assertN("Muchas gracias!", 1, "¡Muchas")
	// CC + no/sí surface after comma (Java surface tokens, not invent)
	assertNTagged("Tengo razón, o no?", map[string]string{"o": "CC"}, 1, "¿o")
	assertN("Tengo razón, no?", 1, "¿no")
	assertN("Tengo razón, eh?", 1, "¿eh")
	assertN("qué me recomendarías???….", 1, "¿qué")
}

func TestQuestionMarkRule_FailClosedWithoutPOS(t *testing.T) {
	rule := NewQuestionMarkRule(nil)
	// Without PT/SPS tags, firstToken stays sentence start word (no question-word invent).
	matches := rule.MatchList(languagetool.AnalyzeTextLocal("Hola, cómo estás?"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "¿Hola", matches[0].GetSuggestedReplacements()[0])
}

func injectESTags(sent *languagetool.AnalyzedSentence, tags map[string]string) {
	if sent == nil {
		return
	}
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok == nil {
			continue
		}
		for surface, pos := range tags {
			if !strings.EqualFold(tok.GetToken(), surface) {
				continue
			}
			p := pos
			tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &p, nil), "test")
		}
	}
}

func formatQM(matches []*rules.RuleMatch) string {
	if len(matches) == 0 {
		return "[]"
	}
	var b strings.Builder
	for i, m := range matches {
		if i > 0 {
			b.WriteString("; ")
		}
		if len(m.GetSuggestedReplacements()) > 0 {
			b.WriteString(m.GetSuggestedReplacements()[0])
		}
	}
	return b.String()
}
