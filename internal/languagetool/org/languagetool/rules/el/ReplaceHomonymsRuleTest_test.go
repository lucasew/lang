package el

// Twin of languagetool-language-modules/el/src/test/java/org/languagetool/rules/el/ReplaceHomonymsRuleTest.java
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

// Java @Before: rule = new ReplaceHomonymsRule(TestTools.getMessages("el"), new Greek());
//               langTool = new JLanguageTool(new Greek());
func replaceHomonymsFixture(t *testing.T) (*ReplaceHomonymsRule, *languagetool.JLanguageTool) {
	t.Helper()
	return NewReplaceHomonymsRule(nil), languagetool.NewJLanguageTool("el")
}

// correct sentences — Java ReplaceHomonymsRuleTest.testRule
func TestReplaceHomonymsRule_Rule(t *testing.T) {
	rule, lt := replaceHomonymsFixture(t)
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("Στην Ελλάδα επικρατεί εύκρατο κλίμα."))))
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("Καλή τύχη σου εύχομαι."))))
}

// test for a wrong usage of a word inside a sentence — Java testRuleInsideOfSentence
func TestReplaceHomonymsRule_RuleInsideOfSentence(t *testing.T) {
	rule, lt := replaceHomonymsFixture(t)
	matches := rule.Match(lt.GetAnalyzedSentence("Του ευχήθηκα καλή τείχη για το διαγώνισμα."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "καλή τύχη", matches[0].GetSuggestedReplacements()[0])
}

// test for a wrong usage of a word in the beggining of a sentence — Java testRuleBegginingOfSentence
// (Java method name typo "Beggining" preserved in Go twin name)
func TestReplaceHomonymsRule_RuleBegginingOfSentence(t *testing.T) {
	rule, lt := replaceHomonymsFixture(t)
	matches := rule.Match(lt.GetAnalyzedSentence("Τεχνητό κόμμα είναι μια ακραία μορφή αναισθησίας."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Τεχνητό κώμα", matches[0].GetSuggestedReplacements()[0])
}

// test for capitalization of suggestion when needed — Java testRuleWithCapitalization
func TestReplaceHomonymsRule_RuleWithCapitalization(t *testing.T) {
	rule, lt := replaceHomonymsFixture(t)
	matches := rule.Match(lt.GetAnalyzedSentence("γάλος πρόεδρος."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Γάλλος πρόεδρος", matches[0].GetSuggestedReplacements()[0])
}

// Field twin of ReplaceHomonymsRule constructor + overrides (beyond Java unit test).
func TestReplaceHomonymsRule_Metadata(t *testing.T) {
	rule := NewReplaceHomonymsRule(nil)

	require.Equal(t, GreekHomonymsReplaceID, rule.GetID())
	require.Equal(t, "GREEK_HOMONYMS_REPLACE", rule.GetID())
	require.Equal(t, "Έλεγχος για λανθασμένη χρήση ομόηχων λέξεων σε μια πρόταση", rule.GetDescription())
	require.Equal(t, "Λανθασμένη χρήση της λέξης", rule.GetShort())
	require.Equal(t, "Μήπως εννοούσατε $suggestions?", rule.GetMessage())
	// Java does not override getSuggestionsSeparator → default ", "
	require.Equal(t, ", ", rule.GetSuggestionsSeparator())
	require.Equal(t, "", rule.SuggestionsSeparator)
	// Java: private static final Locale EL_LOCALE = new Locale("EL");
	require.Equal(t, "EL", rule.GetLocaleCode())
	require.Equal(t, []string{GreekHomonymsFilePath}, rule.GetFileNames())
	require.Equal(t, []string{"/el/replace.txt"}, rule.GetFileNames())
	require.False(t, rule.SubRuleSpecificIDs)
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)

	// Java AbstractSimpleReplaceRule2 ctor: Categories.MISC; Rule default ITS Uncategorized
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "MISC", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSUncategorized, rule.GetLocQualityIssueType())
	// Java does not setUrl / setDefaultOff / addExamplePair
	require.Equal(t, "", rule.GetURL())
	require.False(t, rule.IsDefaultOff())
	require.Nil(t, rule.GetIncorrectExamples())
	require.Nil(t, rule.GetCorrectExamples())

	// Java-visible createMatch message/short for inside-sentence hit
	lt := languagetool.NewJLanguageTool("el")
	matches := rule.Match(lt.GetAnalyzedSentence("Του ευχήθηκα καλή τείχη για το διαγώνισμα."))
	require.Len(t, matches, 1)
	require.Equal(t, "Λανθασμένη χρήση της λέξης", matches[0].GetShortMessage())
	require.Equal(t,
		"Μήπως εννοούσατε <suggestion>καλή τύχη</suggestion>?",
		matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule2
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule2)
	require.True(t, ok, "base rule expected without sub-rule ids")
	require.Equal(t, GreekHomonymsReplaceID, base.GetID())
}
