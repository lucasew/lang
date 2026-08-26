package el

// Twin of languagetool-language-modules/el/src/test/java/org/languagetool/rules/el/GreekRedundancyRuleTest.java
//
// Analysis: Java uses new JLanguageTool(new Greek()).getAnalyzedSentence.
// Go uses NewJLanguageTool("el").GetAnalyzedSentence (same surface/token path).
// AbstractSimpleReplaceRule2 only inspects token surfaces, whitespaceBefore, and
// positions — not POS tags or lemmas.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Java @Before: rule = new GreekRedundancyRule(TestTools.getMessages("el"), new Greek());
//               langTool = new JLanguageTool(new Greek());
func greekRedundancyFixture(t *testing.T) (*GreekRedundancyRule, *languagetool.JLanguageTool) {
	t.Helper()
	return NewGreekRedundancyRule(nil), languagetool.NewJLanguageTool("el")
}

// correct sentences — Java GreekRedundancyRuleTest.testRule
func TestGreekRedundancyRule_Rule(t *testing.T) {
	rule, lt := greekRedundancyFixture(t)
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("Τώρα μπαίνω στο σπίτι."))))
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("Απόψε θα βγω."))))
}

// test for redundancy within the sentence — Java testRuleWithinSentence
func TestGreekRedundancyRule_RuleWithinSentence(t *testing.T) {
	rule, lt := greekRedundancyFixture(t)
	matches := rule.Match(lt.GetAnalyzedSentence("Τώρα μπαίνω μέσα στο σπίτι."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "μπαίνω", matches[0].GetSuggestedReplacements()[0])
}

// test for redundancy in the beggining of a sentence — Java testRuleBegginingOfSentence
func TestGreekRedundancyRule_RuleBegginingOfSentence(t *testing.T) {
	rule, lt := greekRedundancyFixture(t)
	matches := rule.Match(lt.GetAnalyzedSentence("Απόψε το βράδυ θα βγω."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Απόψε", matches[0].GetSuggestedReplacements()[0])
}

// test for redundancy with multiple suggestions — Java testRuleMultipleSuggestions
// File stores a single replacement string containing a comma (not pipe-separated).
func TestGreekRedundancyRule_RuleMultipleSuggestions(t *testing.T) {
	rule, lt := greekRedundancyFixture(t)
	matches := rule.Match(lt.GetAnalyzedSentence("Το μαγαζί ήταν ωραίο, αλλά όμως δεν πέρασα καλά."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "αλλά,όμως", matches[0].GetSuggestedReplacements()[0])
}

// Field twin of GreekRedundancyRule constructor + overrides (beyond Java unit test).
func TestGreekRedundancyRule_Metadata(t *testing.T) {
	rule := NewGreekRedundancyRule(nil)

	require.Equal(t, GreekRedundancyReplaceID, rule.GetID())
	require.Equal(t, "EL_REDUNDANCY_REPLACE", rule.GetID())
	require.Equal(t, "Έλεγχος για χρήση πλεονασμού σε μια πρόταση.", rule.GetDescription())
	require.Equal(t, "Πλεονασμός", rule.GetShort())
	require.Equal(t,
		"'$match' είναι πλεονασμός. Γενικά, είναι προτιμότερο το: $suggestions",
		rule.GetMessage())
	// Java does not override getSuggestionsSeparator → default ", "
	require.Equal(t, ", ", rule.GetSuggestionsSeparator())
	require.Equal(t, "", rule.SuggestionsSeparator)
	require.Equal(t, "el", rule.GetLocaleCode())
	require.Equal(t, []string{GreekRedundancyFilePath}, rule.GetFileNames())
	require.Equal(t, []string{"/el/redundancies.txt"}, rule.GetFileNames())
	require.False(t, rule.SubRuleSpecificIDs)
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "REDUNDANCY", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	// Java does not setUrl
	require.Equal(t, "", rule.GetURL())
	require.False(t, rule.IsDefaultOff())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "Μου αρέσει να <marker>ανεβαίνω πάνω</marker> σε δέντρα.", inc[0].GetExample())
	require.Equal(t, []string{"ανεβαίνω"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "Μου αρέσει να <marker>ανεβαίνω</marker> σε δέντρα.", corr[0].GetExample())

	// Java-visible createMatch message/short for within-sentence hit
	lt := languagetool.NewJLanguageTool("el")
	matches := rule.Match(lt.GetAnalyzedSentence("Τώρα μπαίνω μέσα στο σπίτι."))
	require.Len(t, matches, 1)
	require.Equal(t, "Πλεονασμός", matches[0].GetShortMessage())
	require.Equal(t,
		"'μπαίνω μέσα' είναι πλεονασμός. Γενικά, είναι προτιμότερο το: <suggestion>μπαίνω</suggestion>",
		matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule2
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule2)
	require.True(t, ok, "base rule expected without sub-rule ids")
	require.Equal(t, GreekRedundancyReplaceID, base.GetID())
}
