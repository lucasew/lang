package el

// Twin of languagetool-language-modules/el/src/test/java/org/languagetool/rules/el/GreekSpecificCaseRuleTest.java
//
// Analysis: Java uses JLanguageTool(el).getAnalyzedSentence. Go uses the same
// path (NewJLanguageTool("el").GetAnalyzedSentence) so token spans match Java
// (including double-space toPos=32).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestGreekSpecificCaseRule_Rule(t *testing.T) {
	// Java: new JLanguageTool(Languages.getLanguageForShortCode("el"))
	//       new GreekSpecificCaseRule(TestTools.getEnglishMessages())
	lt := languagetool.NewJLanguageTool("el")
	rule := NewGreekSpecificCaseRule(nil)

	assertGood := func(s string) {
		t.Helper()
		require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence(s))), "good %q", s)
	}
	assertBad := func(s string) []*rules.RuleMatch {
		t.Helper()
		matches := rule.Match(lt.GetAnalyzedSentence(s))
		require.Equal(t, 1, len(matches), "bad %q", s)
		return matches
	}

	// Java GreekSpecificCaseRuleTest.testRule
	assertGood("Ηνωμένες Πολιτείες")
	assertGood("Κατοικώ στις Ηνωμένες Πολιτείες.")
	assertGood("Κατοικώ στις ΗΝΩΜΕΝΕΣ ΠΟΛΙΤΕΙΕΣ.")
	assertBad("ηνωμένες πολιτείες")
	assertBad("ηνωμένες Πολιτείες")
	assertBad("Ηνωμένες πολιτείες")

	matches1 := assertBad("Κατοικώ στις Ηνωμένες πολιτείες.")
	require.Equal(t, 13, matches1[0].GetFromPos())
	require.Equal(t, 31, matches1[0].GetToPos())
	// Java: getSuggestedReplacements().toString() → "[Ηνωμένες Πολιτείες]"
	require.Equal(t, []string{"Ηνωμένες Πολιτείες"}, matches1[0].GetSuggestedReplacements())
	require.Equal(t, "Οι λέξεις της συγκεκριμένης έκφρασης χρείαζεται να ξεκινούν με κεφαλαία γράμματα.", matches1[0].GetMessage())

	// Java: "Κατοικώ στις Ηνωμένες  πολιτείες." (two spaces) → fromPos=13, toPos=32
	matches3 := assertBad("Κατοικώ στις Ηνωμένες  πολιτείες.")
	require.Equal(t, 13, matches3[0].GetFromPos())
	require.Equal(t, 32, matches3[0].GetToPos())
}

// Java GreekSpecificCaseRule ctor metadata: CASING, Misspelling, example pair,
// getPhrasesPath=/el/specific_case.txt, short message.
func TestGreekSpecificCaseRule_Metadata(t *testing.T) {
	rule := NewGreekSpecificCaseRule(nil)
	require.Equal(t, "EL_SPECIFIC_CASE", rule.GetID())
	require.Equal(t, "Ελέγχει αν κάποιες λέξεις χρειάζονται κεφαλαίο το πρώτο τους γράμμα", rule.GetDescription())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "CASING", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())
	// Java does not setUrl
	require.Equal(t, "", rule.GetURL())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "Κατοικώ στις <marker>Ηνωμένες πολιτείες</marker>.", inc[0].GetExample())
	require.Equal(t, []string{"Ηνωμένες Πολιτείες"}, inc[0].GetCorrections())
	require.Equal(t, "Κατοικώ στις <marker>Ηνωμένες Πολιτείες</marker>.", rule.GetCorrectExamples()[0].GetExample())

	// Resource twin of Java getFromResourceDirAsLines("/el/specific_case.txt")
	m, maxLen := loadSpecificCase()
	require.NotEmpty(t, m)
	require.GreaterOrEqual(t, maxLen, 2)
	require.Equal(t, "Ηνωμένες Πολιτείες", m["ηνωμένες πολιτείες"])
	require.Equal(t, "Νέα Υόρκη", m["νέα υόρκη"])

	// Java message branches / short message (exact strings)
	require.Equal(t, "Οι λέξεις της συγκεκριμένης έκφρασης χρείαζεται να ξεκινούν με κεφαλαία γράμματα.", rule.InitialCapitalMessage)
	require.Equal(t, "Η συγκεκριμένη έκφραση γράφεται σύμφωνα με την προτεινόμενη κεφαλαιοποίηση.", rule.OtherCapitalizationMessage)
	require.Equal(t, "Ειδική κεφαλαιοποίηση", rule.ShortMsg)

	// RuleMatch short message from getShortMessage()
	lt := languagetool.NewJLanguageTool("el")
	ms := rule.Match(lt.GetAnalyzedSentence("Ηνωμένες πολιτείες"))
	require.Len(t, ms, 1)
	require.Equal(t, "Ειδική κεφαλαιοποίηση", ms[0].GetShortMessage())
}
