package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/DashRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDashRule_Rule(t *testing.T) {
	rule := NewDashRule(nil)
	assertGood := func(text string) {
		t.Helper()
		require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(text))), "good %q", text)
	}
	assertBad := func(text string) {
		t.Helper()
		require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain(text))), "bad %q", text)
	}

	assertGood("Die große Diäten-Erhöhung kam dann doch.")
	assertGood("Die große Diätenerhöhung kam dann doch.")
	assertGood("Die große Diäten-Erhöhungs-Manie kam dann doch.")
	assertGood("Die große Diäten- und Gehaltserhöhung kam dann doch.")
	assertGood("Die große Diäten- sowie Gehaltserhöhung kam dann doch.")
	assertGood("Die große Diäten- oder Gehaltserhöhung kam dann doch.")
	assertGood("Erst so - Karl-Heinz dann blah.")
	assertGood("Erst so -- Karl-Heinz aber...")
	assertGood("Nord- und Südkorea")
	assertGood("NORD- UND SÜDKOREA")
	assertGood("NORD- BZW. SÜDKOREA")

	assertBad("Die große Diäten- Erhöhung kam dann doch.")
	assertBad("Die große Diäten-  Erhöhung kam dann doch.")
	assertBad("Die große Diäten-Erhöhungs- Manie kam dann doch.")
	assertBad("Die große Diäten- Erhöhungs-Manie kam dann doch.")
	assertBad("MAZEDONIEN- SKOPJE Str.")
}
