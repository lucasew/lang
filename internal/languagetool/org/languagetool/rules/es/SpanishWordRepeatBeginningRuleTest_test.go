package es

// Twin of SpanishWordRepeatBeginningRuleTest — inject RG/LOC_ADV like FreeLing (no surface invent).
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func esWRBMessages() map[string]string {
	return map[string]string{
		"desc_repetition_beginning_adv":       "Tres oraciones sucesivas empiezan con el mismo adverbio.",
		"desc_repetition_beginning_word":      "Tres oraciones sucesivas empiezan con la misma palabra.",
		"desc_repetition_beginning_thesaurus": "Considere usar un diccionario de sinónimos.",
	}
}

func TestSpanishWordRepeatBeginningRule_Rule(t *testing.T) {
	rule := NewSpanishWordRepeatBeginningRule(esWRBMessages())

	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Esto está bien. Esto es mejor, también."))))
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("El coche. El profesor. El tercer elemento."))))

	long := "Por un lado, confirmó que la palabra solo no debe llevar tilde, " +
		"según las reglas generales de acentuación, ni cuando es adverbio, ni cuando es adjetivo. Por otro lado, y este " +
		"es el tema que hoy nos interesa, confirmó que los demostrativos este, ese o aquel, y sus formas femeninas y " +
		"plurales, no llevarán tampoco tilde funcionando tanto como pronombres como determinantes."
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze(long))))

	// Mañana: Java FreeLing RG POS — inject
	require.Equal(t, 2, len(rule.MatchList(analyzeESWRBSentences("Mañana me voy. Mañana vuelvo. Mañana se acabó todo."))))

	// without RG, Mañana is not isAdverb; only third-sentence word path may still fire once
	require.Equal(t, 1, len(rule.MatchList(languagetool.SplitAndAnalyze("Mañana me voy. Mañana vuelvo. Mañana se acabó todo."))))

	matches1 := rule.MatchList(languagetool.SplitAndAnalyze("Yo creo. Yo he visto esto antes. Yo no lo creo."))
	require.Equal(t, 1, len(matches1))
	require.Equal(t, "Además, yo", matches1[0].GetSuggestedReplacements()[0])
	require.Equal(t, "Igualmente, yo", matches1[0].GetSuggestedReplacements()[1])
	require.Equal(t, "No solo eso, sino que yo", matches1[0].GetSuggestedReplacements()[2])

	matches2 := rule.MatchList(languagetool.SplitAndAnalyze("También, juego a fútbol. También, juego a baloncesto."))
	require.Equal(t, 1, len(matches2))
	require.Equal(t, "[Adicionalmente, Asimismo, Además, Igualmente, Así mismo]",
		formatSugg(matches2[0].GetSuggestedReplacements()))

	// Sin embargo: LOC_ADV inject on "Sin" (multiword chunk in Java)
	matches3 := rule.MatchList(analyzeESWRBSentences("Sin embargo, me gusta. Sin embargo, otros me gustan más."))
	require.Equal(t, 1, len(matches3))
	require.Equal(t, "[]", formatSugg(matches3[0].GetSuggestedReplacements()))

	matches4 := rule.MatchList(languagetool.SplitAndAnalyze("Pero me gusta. Pero otros me gustan más."))
	require.Equal(t, 1, len(matches4))
	require.Equal(t, "[Aun así, Por otra parte, Sin embargo]", formatSugg(matches4[0].GetSuggestedReplacements()))
}

func analyzeESWRBSentences(text string) []*languagetool.AnalyzedSentence {
	parts := languagetool.SplitAndAnalyze(text)
	for _, s := range parts {
		for _, tok := range s.GetTokensWithoutWhitespace() {
			if tok == nil || tok.IsSentenceStart() {
				continue
			}
			// first content word only
			t := tok.GetToken()
			if strings.EqualFold(t, "mañana") {
				pos := "RG"
				tok.AddReading(languagetool.NewAnalyzedToken(t, &pos, nil), "test")
			}
			if t == "Sin" {
				pos := "LOC_ADV"
				tok.AddReading(languagetool.NewAnalyzedToken(t, &pos, nil), "test")
			}
			break
		}
	}
	return parts
}

func formatSugg(s []string) string {
	return "[" + joinComma(s) + "]"
}

func joinComma(s []string) string {
	if len(s) == 0 {
		return ""
	}
	out := s[0]
	for i := 1; i < len(s); i++ {
		out += ", " + s[i]
	}
	return out
}
