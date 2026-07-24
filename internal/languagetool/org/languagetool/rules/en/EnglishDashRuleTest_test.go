package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishDashRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Java AbstractDashRule.match only uses sentence.getText(); AnalyzePlain is enough
// (same as JLanguageTool(en-US).getAnalyzedSentence for this rule).
func TestEnglishDashRule_Rule(t *testing.T) {
	rule := NewEnglishDashRule(nil)
	check := func(expectedErrors int, text string, expectedSuggestion string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(text))
		require.Equal(t, expectedErrors, len(matches), "text %q got %v", text, matches)
		if expectedSuggestion != "" && len(matches) > 0 {
			require.Equal(t, "A dash was used instead of a hyphen.", matches[0].GetMessage())
			require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
			require.Equal(t, expectedSuggestion, matches[0].GetSuggestedReplacements()[0])
		}
	}

	// correct sentences (Java EnglishDashRuleTest):
	check(0, "This is my T-shirt.", "")
	check(0, "This is water-proof.", "")
	check(0, "This works semi-automatically.", "")
	check(0, "She's a newcomer.", "")   // compounds.txt: "new-comer+"
	check(0, "I sent you and e-mail.", "") // compounds.txt: "e-mail"
	// incorrect sentences:
	check(1, "T – shirt", "T-shirt")
	check(1, "three–way street", "three-way")
	check(1, "surface — to — surface", "surface-to-surface")
	check(1, "This works semi–automatically.", "semi-automatically") // semi-automatic$ / semi-automatically$
	check(1, "This works semi – automatically.", "semi-automatically")
	check(1, "I sent you and e–mail.", "e-mail")
	// Java: new-comer hyphen→closed form not yet supported (commented out in Java)
}

// Java EnglishDashRule ctor: ID, description, message, hyphen URL, example pair, TYPOGRAPHY + picky.
func TestEnglishDashRule_Metadata(t *testing.T) {
	rule := NewEnglishDashRule(nil)
	require.Equal(t, "EN_DASH_RULE", rule.GetID())
	require.Equal(t, "Checks if hyphenated words were spelled with dashes (e.g., 'T — shirt' instead 'T-shirt').", rule.GetDescription())
	require.Equal(t, "A dash was used instead of a hyphen.", rule.GetMessage())
	require.Equal(t, "https://languagetool.org/insights/post/hyphen/", rule.GetURL())
	require.Equal(t, 2, rule.EstimateContextForSureMatch())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "TYPOGRAPHY", rule.GetCategory().GetID().String())
	require.True(t, rule.HasTag(rules.TagPicky))
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "I'll buy a new <marker>T—shirt</marker>.", inc[0].GetExample())
	require.Equal(t, []string{"T-shirt"}, inc[0].GetCorrections())
	require.Equal(t, "I'll buy a new <marker>T-shirt</marker>.", rule.GetCorrectExamples()[0].GetExample())
}
