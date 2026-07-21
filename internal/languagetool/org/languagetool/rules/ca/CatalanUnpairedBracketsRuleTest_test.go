package ca

// Twin of languagetool-language-modules/ca/src/test/java/org/languagetool/rules/ca/CatalanUnpairedBracketsRuleTest.java
// Reduced surface cases (full Java has many exceptions).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCatalanUnpairedBracketsRule_Rule(t *testing.T) {
	rule := NewCatalanUnpairedBracketsRule(nil)
	matchN := func(s string) int {
		return len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}))
	}
	require.Equal(t, 0, matchN("Això és (correcte)."))
	require.Equal(t, 1, matchN("Això és (incorrecte."))
	require.Equal(t, 1, matchN("Això és incorrecte)."))
}

// Twin of CatalanUnpairedBracketsRuleTest.testMultipleSentences
func TestCatalanUnpairedBracketsRule_MultipleSentences(t *testing.T) {
	rule := NewCatalanUnpairedBracketsRule(nil)
	matchN := func(sents ...string) int {
		var as []*languagetool.AnalyzedSentence
		for _, s := range sents {
			as = append(as, languagetool.AnalyzePlain(s))
		}
		return len(rule.MatchList(as))
	}
	// paired brackets across sentences → 0
	require.Equal(t, 0, matchN(
		"Aquesta és una sentència múltiple amb claudàtors: [Ací hi ha un claudàtor.",
		"Amb algun text.] i ací continua.\n",
	))
	require.Equal(t, 0, matchN("\"Era la teva filla.", "El corcó no et rosegarà més.\"\n\n"))
	require.Equal(t, 0, matchN("\"Era la teva filla.", "El corcó no et rosegarà més\".\n\n"))
	// unclosed [
	require.Equal(t, 1, matchN(
		"Aquesta és una sentència múltiple amb claudàtors: [Ací hi ha un claudàtor.",
		"Amb algun text. I ací continua.\n\n",
	))
	// parentheses across blank line
	require.Equal(t, 0, matchN(
		"Aquesta és una sentència múltiple amb parèntesis (Ací hi ha un parèntesi.",
		"\n\n Amb algun text.) i ací continua.",
	))
}

// Twin of CatalanUnpairedBracketsRuleTest.testQuestionExclamation
func TestCatalanUnpairedBracketsRule_QuestionExclamation(t *testing.T) {
	// Java enables CA_UNPAIRED_QUESTION / CA_UNPAIRED_EXCLAMATION (not brackets).
	// Catalan tokenizer splits leading dialogue dash from the word; AnalyzePlain keeps "-Com".
	q := NewCatalanUnpairedQuestionMarksRule(nil)
	e := NewCatalanUnpairedExclamationMarksRule(nil)

	caSent := func(toks ...string) *languagetool.AnalyzedSentence {
		var atrs []*languagetool.AnalyzedTokenReadings
		ss := languagetool.SentenceStartTagName
		atrs = append(atrs, languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0))
		pos := 0
		for _, w := range toks {
			atrs = append(atrs, languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(w, nil, nil), pos))
			pos += len([]rune(w)) + 1
		}
		return languagetool.NewAnalyzedSentence(atrs)
	}

	// "- Com estàs ?" with separate dash token (Java)
	ms := q.MatchList([]*languagetool.AnalyzedSentence{caSent("-", "Com", "estàs", "?")})
	require.Equal(t, 1, len(ms))
	require.Equal(t, "¿Com", ms[0].GetSuggestedReplacements()[0])

	ms = e.MatchList([]*languagetool.AnalyzedSentence{caSent("-", "Benvinguda", "!")})
	require.Equal(t, 1, len(ms))
	require.Equal(t, "¡Benvinguda", ms[0].GetSuggestedReplacements()[0])

	ms = q.MatchList([]*languagetool.AnalyzedSentence{caSent("-", "Tinc", "raó", ",", "oi", "?")})
	require.Equal(t, 1, len(ms))
	require.Equal(t, "¿oi", ms[0].GetSuggestedReplacements()[0])
}

