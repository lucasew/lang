package el

// Twin of languagetool-language-modules/el/src/test/java/org/languagetool/rules/el/NumeralStressRuleTest.java
//
// Analysis: Java uses new JLanguageTool(new Greek()).getAnalyzedSentence.
// Go uses NewJLanguageTool("el").GetAnalyzedSentence so tokens match the Greek path.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Java NumeralStressRuleTest.testRule
func TestNumeralStressRule_Rule(t *testing.T) {
	// Java: rule = new NumeralStressRule(TestTools.getMessages("el"));
	//       lt = new JLanguageTool(new Greek());
	rule := NewNumeralStressRule(nil)
	lt := languagetool.NewJLanguageTool("el")

	assertCorrect := func(sentence string) {
		t.Helper()
		matches := rule.Match(lt.GetAnalyzedSentence(sentence))
		require.Equal(t, 0, len(matches), "correct %q", sentence)
	}
	assertIncorrect := func(sentence, correction string) {
		t.Helper()
		matches := rule.Match(lt.GetAnalyzedSentence(sentence))
		require.Equal(t, 1, len(matches), "incorrect %q", sentence)
		require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()), sentence)
		require.Equal(t, correction, matches[0].GetSuggestedReplacements()[0], sentence)
	}

	assertCorrect("1ος")
	assertCorrect("2η")
	assertCorrect("3ο")
	assertCorrect("20ός")
	assertCorrect("30ή")
	assertCorrect("40ό")
	assertCorrect("1000ών")
	assertCorrect("1010ες")

	assertIncorrect("4ός", "4ος")
	assertIncorrect("5ή", "5η")
	assertIncorrect("6ό", "6ο")
	assertIncorrect("100ος", "100ός")
	assertIncorrect("200η", "200ή")
	assertIncorrect("300ο", "300ό")
	assertIncorrect("2000ων", "2000ών")
	assertIncorrect("2010ές", "2010ες")
	assertIncorrect("2020α", "2020ά")
}

// Java NumeralStressRule ctor metadata: ORTHOGRAPHY, Misspelling, example pair, description.
func TestNumeralStressRule_Metadata(t *testing.T) {
	rule := NewNumeralStressRule(nil)
	require.Equal(t, "GREEK_ORTHOGRAPHY_NUMERAL_STRESS", rule.GetID())
	require.Equal(t, "Έλεγχος τονισμού αριθμητικών", rule.GetDescription())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "ORTHOGRAPHY", rule.GetCategory().GetID().String())
	require.Equal(t, "Orthography", rule.GetCategory().GetName())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "Ο <marker>20ος</marker> αιώνας μαζί με τον 21ο αιώνα κατατάσσεται από τους ιστορικούς στη Σύγχρονη Ιστορία.", inc[0].GetExample())
	// Java IncorrectExample from fixed <marker> span
	require.Equal(t, []string{"20ός"}, inc[0].GetCorrections())
	require.Equal(t, "Ο <marker>20ός</marker> αιώνας μαζί με τον 21ο αιώνα κατατάσσεται από τους ιστορικούς στη Σύγχρονη Ιστορία.", rule.GetCorrectExamples()[0].GetExample())

	// Message / short message from a real match (Java RuleMatch msg + shortMessage)
	lt := languagetool.NewJLanguageTool("el")
	ms := rule.Match(lt.GetAnalyzedSentence("20ος"))
	require.Len(t, ms, 1)
	require.Equal(t, "<suggestion>20ός</suggestion>", ms[0].GetMessage())
	require.Equal(t, "Πρόβλημα ορθογραφίας", ms[0].GetShortMessage())
	require.Equal(t, []string{"20ός"}, ms[0].GetSuggestedReplacements())
}
