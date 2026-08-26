package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/NewZealandReplaceRuleTest.java
//
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START) is equivalent to Java
// JLanguageTool(NewZealandEnglish).getAnalyzedSentence for this rule’s surface matching.
// NewZealandReplaceRule/AbstractSimpleReplaceRule2 only inspects token surfaces,
// whitespaceBefore, and positions — not POS tags or lemmas. Official replace.txt
// key "sidewalk" tokenizes the same under WordTokenizer as EN.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Twin of NewZealandReplaceRuleTest.testRule
func TestNewZealandReplaceRule_Rule(t *testing.T) {
	// Java: new NewZealandReplaceRule(TestTools.getMessages("en"), "/en/en-NZ/replace.txt")
	//       lt = new JLanguageTool(new NewZealandEnglish())
	rule := NewNewZealandReplaceRule(nil, NewZealandReplaceFilePath)

	// correct sentences:
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Walk on the footpath."))))

	// incorrect sentences:

	// at the beginning of a sentence (Romanian replace rule is case-sensitive)
	checkNZSimpleReplaceRule(t, rule, "Sidewalk is not a place to park your car!", "Footpath")
	// inside sentence
	checkNZSimpleReplaceRule(t, rule, "I walked on the sidewalk", "footpath")

	// Java-visible createMatch extras (no useSubRuleSpecificIds — match rule is the base rule)
	matches := rule.Match(languagetool.AnalyzePlain("I walked on the sidewalk"))
	require.Len(t, matches, 1)
	require.Equal(t, "Not a New Zealand English word", matches[0].GetShortMessage())
	require.Equal(t,
		"'sidewalk' is a non-standard expression. Consider using expressions more common to New Zealand English.",
		matches[0].GetMessage())
	// Java does NOT call useSubRuleSpecificIds — RuleMatch.getRule() is the base rule.
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "SubRuleSpecificIDs must stay false (Java does not enable it)")
	baseRule, ok := matches[0].GetRule().(rules.Rule)
	require.True(t, ok, "match rule must implement rules.Rule")
	require.Equal(t, NewZealandSimpleReplaceRuleID, baseRule.GetID())
	require.Equal(t, "EN_NZ_SIMPLE_REPLACE", baseRule.GetID())
	require.Equal(t, "English words easily confused in New Zealand English", baseRule.GetDescription())
	// "I walked on the sidewalk" — sidewalk starts after "I walked on the "
	require.Equal(t, 16, matches[0].GetFromPos())
	require.Equal(t, 24, matches[0].GetToPos())
}

// checkNZSimpleReplaceRule ports NewZealandReplaceRuleTest.checkSimpleReplaceRule.
func checkNZSimpleReplaceRule(t *testing.T, rule *NewZealandReplaceRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()),
		"Invalid replacement count wile checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
}

// Field twin of NewZealandReplaceRule constructor + overrides (not in Java test class,
// but required for sector field-by-field parity).
func TestNewZealandReplaceRule_Metadata(t *testing.T) {
	rule := NewNewZealandReplaceRule(nil, NewZealandReplaceFilePath)

	require.Equal(t, NewZealandSimpleReplaceRuleID, rule.GetID())
	require.Equal(t, "EN_NZ_SIMPLE_REPLACE", rule.GetID())
	require.Equal(t, "English words easily confused in New Zealand English", rule.GetDescription())
	require.Equal(t, "Not a New Zealand English word", rule.GetShort())
	require.Equal(t, "'$match' is a non-standard expression. Consider using expressions more common to New Zealand English.", rule.GetMessage())
	require.Equal(t, "en-NZ", rule.GetLocaleCode())
	require.Equal(t, []string{NewZealandReplaceFilePath}, rule.GetFileNames())
	require.Equal(t, []string{"/en/en-NZ/replace.txt"}, rule.GetFileNames())
	require.False(t, rule.SubRuleSpecificIDs, "Java NewZealandReplaceRule does not call useSubRuleSpecificIds()")
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "STYLE", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSLocaleViolation, rule.GetLocQualityIssueType())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "A <marker>sidewalk</marker> is a path along the side of a road.", inc[0].GetExample())
	require.Equal(t, []string{"footpath"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "A <marker>footpath</marker> is a path along the side of a road.", corr[0].GetExample())
}
