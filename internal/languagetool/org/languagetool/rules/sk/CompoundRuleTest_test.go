package sk

// Twin of languagetool-language-modules/sk/src/test/java/org/languagetool/rules/sk/CompoundRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestCompoundRule_Metadata(t *testing.T) {
	r := NewCompoundRule(nil)
	require.Equal(t, "SK_COMPOUNDS", r.GetID())
	require.Equal(t, "Slová so spojovníkom napr. použite „česko-slovenský” namiesto „česko slovenský”", r.GetDescription())
	// Java AbstractCompoundRule ctor: Categories.MISC + ITSIssueType.Misspelling
	require.Equal(t, rules.CatMisc.GetCategory(nil).GetID(), r.GetCategory().GetID())
	require.Equal(t, rules.ITSMisspelling, r.GetLocQualityIssueType())
	// Java 4-arg super shortMessage
	require.Equal(t, "Problém spájania slov", r.ShortDesc)
	require.Equal(t, "Toto slovo sa zvyčajne píše so spojovníkom.", r.WithHyphenMessage)
	require.Equal(t, "Toto slovo sa obvykle píše bez spojovníka.", r.WithoutHyphenMessage)
	require.Equal(t, "Tento výraz sa bežne píše s alebo bez spojovníka.", r.WithOrWithoutHyphenMessage)
	// Java AbstractCompoundRule default / field true
	require.True(t, r.SentenceStartsWithUpperCase)
	// Java getCompoundRuleData singleton from /sk/compounds.txt
	require.NotNil(t, r.GetCompoundRuleData())
	require.Same(t, loadCompoundData(), r.GetCompoundRuleData())
}

func TestCompoundRule_Rule(t *testing.T) {
	rule := NewCompoundRule(nil)
	check := func(expectedErrors int, text string, expSuggestions ...string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(text))
		require.Equal(t, expectedErrors, len(matches), "text %q got %v", text, matches)
		if len(expSuggestions) > 0 {
			require.Equal(t, 1, expectedErrors)
			require.Equal(t, expSuggestions, matches[0].GetSuggestedReplacements())
		}
	}
	// correct:
	check(0, "česko-slovenský")
	check(0, "rakúsko-uhorský")
	// incorrect:
	check(1, "bim bam", "bim-bam")
}
