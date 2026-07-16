package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/WordCoherencyRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func analyzeDECoherency(s string) []*languagetool.AnalyzedSentence {
	return languagetool.AnalyzeTextLocal(s)
}

func TestWordCoherencyRule_Rule(t *testing.T) {
	assertGood := func(s string) {
		t.Helper()
		rule := NewWordCoherencyRule(nil)
		require.Equal(t, 0, len(rule.MatchList(analyzeDECoherency(s))), "good: %q", s)
	}
	assertError := func(s string, expectedSuggestion ...string) {
		t.Helper()
		rule := NewWordCoherencyRule(nil)
		matches := rule.MatchList(analyzeDECoherency(s))
		require.Equal(t, 1, len(matches), "error: %q got %d", s, len(matches))
		if len(expectedSuggestion) > 0 {
			require.Equal(t, expectedSuggestion[0], matches[0].GetSuggestedReplacements()[0], "sugg for %q", s)
		}
	}

	assertGood("Das ist aufwendig, aber nicht zu aufwendig.")
	assertGood("Das ist aufwendig. Aber nicht zu aufwendig.")
	assertGood("Das ist aufwändig, aber nicht zu aufwändig.")
	assertGood("Das ist aufwändig. Aber nicht zu aufwändig.")

	assertError("Das ist aufwendig. Aufwändig ist das.", "Aufwendig")
	assertError("Das ist aufwendig, aber nicht zu aufwändig.")
	assertError("Das ist aufwendig. Aber nicht zu aufwändig.")
	assertError("Das ist aufwendiger, aber nicht zu aufwändig.")
	assertError("Das ist aufwendiger. Aber nicht zu aufwändig.")
	assertError("Das ist aufwändig, aber nicht zu aufwendig.")
	assertError("Das ist aufwändig. Aber nicht zu aufwendig.")
	assertError("Das ist aufwändiger, aber nicht zu aufwendig.")
	assertError("Das ist aufwändiger. Aber nicht zu aufwendig.")
	assertError("Delfin und Delphin")
	assertError("Delfins und Delphine")
	assertError("essentiell und essenziell")
	assertError("essentieller und essenzielles")
	assertError("Differential und Differenzial")
	assertError("Differentials und Differenzials")
	assertError("Facette und Fassette")
	assertError("Facetten und Fassetten")
	assertError("Joghurt und Jogurt")
	assertError("Joghurts und Jogurt")
	assertError("Joghurt und Jogurts")
	assertError("Joghurts und Jogurts")
	assertError("Ketchup und Ketschup")
	assertError("Ketchups und Ketschups")
	assertError("Kommuniqué und Kommunikee")
	assertError("Kommuniqués und Kommunikees")
	assertError("Necessaire und Nessessär")
	assertError("Necessaires und Nessessärs")
	assertError("Orthographie und Orthografie")
	assertError("substantiell und substanziell")
	assertError("substantieller und substanzielles")
	assertError("Thunfisch und Tunfisch")
	assertError("Thunfische und Tunfische")
	assertError("Xylophon und Xylofon")
	assertError("Xylophone und Xylofone")
	assertError("selbständig und selbstständig")
	assertError("selbständiges und selbstständiger")
	assertError("Bahnhofsplatz und Bahnhofplatz")

	rule := NewWordCoherencyRule(nil)
	matches1 := rule.MatchList(analyzeDECoherency("Eine aufwendige Untersuchung. Oder ist sie aufwändig?"))
	require.Equal(t, 1, len(matches1))
	require.Equal(t, 43, matches1[0].GetFromPos())
	require.Equal(t, 52, matches1[0].GetToPos())
	require.Equal(t, []string{"aufwendig"}, matches1[0].GetSuggestedReplacements())

	matches2 := rule.MatchList(analyzeDECoherency("Eine aufwendige Untersuchung. Oder ist sie noch aufwändiger?"))
	require.Equal(t, 1, len(matches2))
	require.Equal(t, 48, matches2[0].GetFromPos())
	require.Equal(t, 59, matches2[0].GetToPos())
	require.Equal(t, []string{"aufwendiger"}, matches2[0].GetSuggestedReplacements())
}

func TestWordCoherencyRule_CallIndependence(t *testing.T) {
	assertGood := func(s string) {
		t.Helper()
		require.Equal(t, 0, len(NewWordCoherencyRule(nil).MatchList(analyzeDECoherency(s))))
	}
	assertGood("Das ist aufwendig.")
	assertGood("Aber nicht zu aufwändig.")
}

func TestWordCoherencyRule_MatchPosition(t *testing.T) {
	matches := NewWordCoherencyRule(nil).MatchList(analyzeDECoherency("Das ist aufwendig. Aber nicht zu aufwändig"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 33, matches[0].GetFromPos())
	require.Equal(t, 42, matches[0].GetToPos())
}

func TestWordCoherencyRule_RuleCompleteTexts(t *testing.T) {
	check := func(s string) int {
		return len(NewWordCoherencyRule(nil).MatchList(analyzeDECoherency(s)))
	}
	require.Equal(t, 0, check("Das ist aufwändig. Aber hallo. Es ist wirklich aufwändig."))
	require.Equal(t, 1, check("Das ist aufwendig. Aber hallo. Es ist wirklich aufwändig."))
	require.Equal(t, 1, check("Das ist aufwändig. Aber hallo. Es ist wirklich aufwendig."))
	require.Equal(t, 0, check("Das ist aufwendig. Aber hallo. Es ist wirklich aufwendiger als so."))
	require.Equal(t, 1, check("Das ist aufwendig. Aber hallo. Es ist wirklich aufwändiger als so."))
	require.Equal(t, 1, check("Das ist aufwändig. Aber hallo. Es ist wirklich aufwendiger als so."))
	require.Equal(t, 1, check("Das ist das aufwändigste. Aber hallo. Es ist wirklich aufwendiger als so."))
	require.Equal(t, 1, check("Das ist das aufwändigste. Aber hallo. Es ist wirklich aufwendig."))
	// cross-paragraph: AnalyzeTextLocal may not split on \n\n alone — use AnalyzeTextDemo if needed
	require.Equal(t, 1, check("Das ist das aufwändigste.\n\nAber hallo. Es ist wirklich aufwendig."))
}
