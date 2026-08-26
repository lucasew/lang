package ro

// Twin of languagetool-language-modules/ro/src/test/java/org/languagetool/rules/ro/CompoundRuleTest.java
//
// Java setUp:
//   lt = new JLanguageTool(new Romanian());
//   rule = new CompoundRule(TestTools.getMessages("ro"), new Romanian(), null);
// Java testRule → AbstractCompoundRuleTest.check(expected, text, suggestions…)
// via rule.match(lt.getAnalyzedSentence(text)).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestCompoundRule_Metadata(t *testing.T) {
	r := NewCompoundRule(nil)
	// Java getId / getDescription
	require.Equal(t, "RO_COMPOUND", r.GetID())
	require.Equal(t, "Greșeală de scriere (cuvinte scrise legat sau cu cratimă)", r.GetDescription())
	// Java AbstractCompoundRule ctor: Categories.MISC + ITSIssueType.Misspelling
	require.Equal(t, rules.CatMisc.GetCategory(nil).GetID(), r.GetCategory().GetID())
	require.Equal(t, rules.ITSMisspelling, r.GetLocQualityIssueType())
	// Java 4-arg super shortMessage + RO messages
	require.Equal(t, "Problemă de scriere (cratimă, spațiu, etc.)", r.ShortDesc)
	require.Equal(t, "Cuvântul se scrie cu cratimă.", r.WithHyphenMessage)
	require.Equal(t, "Cuvântul se scrie legat.", r.WithoutHyphenMessage)
	require.Equal(t, "Cuvântul se scrie legat sau cu cratimă.", r.WithOrWithoutHyphenMessage)
	// Java AbstractCompoundRule default / field true
	require.True(t, r.SentenceStartsWithUpperCase)
	// Java getCompoundRuleData singleton from /ro/compounds.txt
	require.NotNil(t, r.GetCompoundRuleData())
	require.Same(t, loadCompoundData(), r.GetCompoundRuleData())
}

func TestCompoundRule_Rule(t *testing.T) {
	// Java: lt = new JLanguageTool(new Romanian()); rule = new CompoundRule(...);
	lt := languagetool.NewJLanguageTool("ro")
	rule := NewCompoundRule(nil)

	check := func(expectedErrors int, text string, expSuggestions ...string) {
		t.Helper()
		// Java AbstractCompoundRuleTest.check → rule.match(lt.getAnalyzedSentence(text))
		sent := lt.GetAnalyzedSentence(text)
		require.NotNil(t, sent, "GetAnalyzedSentence(%q)", text)
		matches := rule.Match(sent)
		require.Equal(t, expectedErrors, len(matches), "text %q got %v", text, matches)
		if len(expSuggestions) > 0 {
			require.Equal(t, 1, expectedErrors)
			require.Equal(t, expSuggestions, matches[0].GetSuggestedReplacements(), "text %q", text)
			// compounds.txt entries end with '+' → joined-only → WithoutHyphenMessage
			require.Equal(t, "Cuvântul se scrie legat.", matches[0].GetMessage(), "text %q", text)
			require.Equal(t, "Problemă de scriere (cratimă, spațiu, etc.)", matches[0].GetShortMessage(), "text %q", text)
		}
	}

	// correct sentences:
	check(0, "Au plecat câteșitrei.")
	// incorrect sentences:
	check(1, "câte și trei", "câteșitrei")
	check(1, "Câte și trei", "Câteșitrei")
	// check(1, "Câte și trei", new String[] { "Câteșitrei" }); // Java duplicate commented
	check(1, "câte-și-trei", "câteșitrei")

	check(1, "tus trei", "tustrei")
	check(1, "tus-trei", "tustrei")
}
