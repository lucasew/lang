package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/DashRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// TestDashRule_Rule ports DashRuleTest.testRule.
// Java: JLanguageTool(Languages.getLanguageForShortCode("de-DE")).getAnalyzedSentence(text)
// then rule.match(...).length == 0 (good) or 1 (bad). Match only reads
// getTokensWithoutWhitespace surfaces/positions — morph tags unused.
func TestDashRule_Rule(t *testing.T) {
	rule := NewDashRule(nil) // Java: new DashRule(TestTools.getMessages("de"))
	lt := languagetool.NewJLanguageTool("de-DE")

	assertGood := func(text string) {
		t.Helper()
		require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence(text))), "good %q", text)
	}
	assertBad := func(text string) {
		t.Helper()
		require.Equal(t, 1, len(rule.Match(lt.GetAnalyzedSentence(text))), "bad %q", text)
	}

	// correct sentences (Java DashRuleTest):
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

	// incorrect sentences:
	assertBad("Die große Diäten- Erhöhung kam dann doch.")
	assertBad("Die große Diäten-  Erhöhung kam dann doch.")
	assertBad("Die große Diäten-Erhöhungs- Manie kam dann doch.")
	assertBad("Die große Diäten- Erhöhungs-Manie kam dann doch.")
	assertBad("MAZEDONIEN- SKOPJE Str.")
}

// Extra outcome checks beyond Java assert length: suggestions + meta surface of Rule.
func TestDashRule_SuggestionsAndMeta(t *testing.T) {
	rule := NewDashRule(nil)
	lt := languagetool.NewJLanguageTool("de-DE")

	// Java match: join always; comma form only when total '-' count ≤ 1
	ms := rule.Match(lt.GetAnalyzedSentence("Die große Diäten- Erhöhung kam dann doch."))
	require.Equal(t, 1, len(ms))
	require.Equal(t, []string{"Diäten-Erhöhung", "Diäten-, Erhöhung"}, ms[0].GetSuggestedReplacements())
	require.Equal(t, "Fehlendes 'und' oder Komma oder überflüssiges Leerzeichen?", ms[0].ShortMessage)
	require.Contains(t, ms[0].GetMessage(), "überflüssiges Leerzeichen")

	// two hyphens on left token → only concat suggestion
	ms2 := rule.Match(lt.GetAnalyzedSentence("Die große Diäten-Erhöhungs- Manie kam dann doch."))
	require.Equal(t, 1, len(ms2))
	require.Equal(t, []string{"Diäten-Erhöhungs-Manie"}, ms2[0].GetSuggestedReplacements())

	// Java equalsAny exact "UND"|"ODER"|"BZW" only; tokenizer splits "BZW." → "BZW" + "."
	s := lt.GetAnalyzedSentence("NORD- BZW. SÜDKOREA")
	require.Equal(t, 0, len(rule.Match(s)))

	require.Equal(t, "DE_DASH", rule.GetID())
	require.Equal(t, "Keine Leerzeichen in Bindestrich-Komposita (wie z.B. in 'Diäten- Erhöhung')", rule.GetDescription())
	require.Contains(t, rule.GetURL(), "grammatik-leerzeichen")
	require.NotNil(t, rule.GetCategory())
	require.Len(t, rule.GetIncorrectExamples(), 1)
	require.Len(t, rule.GetCorrectExamples(), 1)
	require.Equal(t, []string{"Diäten-Erhöhung"}, rule.GetIncorrectExamples()[0].GetCorrections())
}
