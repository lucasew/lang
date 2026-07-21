package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/DuUpperLowerCaseRuleTest.java
// Java uses singletonList(getAnalyzedSentence(input)) — whole string as one analyzed sentence.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDuUpperLowerCaseRule_Rule(t *testing.T) {
	rule := NewDuUpperLowerCaseRule(nil)
	assertErrors := func(input string, expected int) {
		t.Helper()
		// Java: rule.match(Collections.singletonList(lt.getAnalyzedSentence(input)))
		n := len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(input)}))
		require.Equal(t, expected, n, "input=%q", input)
	}

	// correct (0):
	assertErrors("Du bist noch jung.", 0)
	assertErrors("Du bist noch jung, du bist noch fit.", 0)
	assertErrors("Aber du bist noch jung, du bist noch fit.", 0)
	assertErrors("Aber du bist noch jung, dir ist das egal.", 0)
	assertErrors("Hast Du ihre Brieftasche gesehen?", 0)

	// mixed Du/du errors (1):
	assertErrors("Aber Du bist noch jung, du bist noch fit.", 1)
	assertErrors("Aber Du bist noch jung, dir ist das egal.", 1)
	assertErrors("Aber Du bist noch jung. Und dir ist das egal.", 1)

	assertErrors("Aber du bist noch jung. Und Du bist noch fit.", 1)
	assertErrors("Aber du bist noch jung, Dir ist das egal.", 1)
	assertErrors("Aber du bist noch jung. Und Dir ist das egal.", 1)

	// euer / Euer coherence:
	assertErrors("Aber du bist noch jung, sagt euer Vater oft.", 0)
	assertErrors("Aber Du bist noch jung, sagt Euer Vater oft.", 0)
	assertErrors("Aber Du bist noch jung, sagt euer Vater oft.", 1)
	assertErrors("Aber du bist noch jung, sagt Euer Vater oft.", 1)

	// Ihr / ihr:
	assertErrors("Könnt Ihr Euch das vorstellen???", 0)
	assertErrors("Könnt ihr euch das vorstellen???", 0)
	assertErrors("Aber Samstags geht ihr Sohn zum Sport. Stellt Euch das mal vor!", 0)
	// commented out in Java:
	// assertErrors("Könnt Ihr euch das vorstellen???", 1)
	assertErrors("Wie geht es euch? Herr Meier, wie war ihr Urlaub?", 0)
	assertErrors("Wie geht es Euch? Herr Meier, wie war Ihr Urlaub?", 0)

	// more goods:
	assertErrors("\"Du sagtest, du würdest es schaffen!\"", 0)
	assertErrors("Egal, was du tust: Du musst dein Bestes geben.", 0)
	assertErrors("Was auch immer du tust: ICH UND DU KÖNNEN ES SCHAFFEN.", 0)
	assertErrors("Hast Du die Adresse von ihr?", 0)

	// Class example (Dir then du) — multi-sentence as two list items also used in Go
	// when AnalyzePlain of two sentences in one string may still work as one sentence token stream.
	assertErrors("Wie geht es Dir? Bist du wieder gesund?", 1)

	require.Equal(t, "https://languagetool.org/insights/de/beitrag/duzen-grossgeschrieben/", rule.GetURL())
	require.Equal(t, -1, rule.MinToCheckParagraph())
	// Java: return "DE_DU_UPPER_LOWER";
	require.Equal(t, "DE_DU_UPPER_LOWER", rule.GetID())
}
