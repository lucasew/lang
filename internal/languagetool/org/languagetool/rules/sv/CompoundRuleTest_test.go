package sv

// Twin of languagetool-language-modules/sv/src/test/java/org/languagetool/rules/sv/CompoundRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestCompoundRule_Metadata(t *testing.T) {
	r := NewCompoundRule(nil)
	require.Equal(t, "SV_COMPOUNDS", r.GetID())
	require.Equal(t, "Särskrivningar, t.ex. 'e mail' bör skrivas 'e-mail'", r.GetDescription())
	// Java AbstractCompoundRule ctor: Categories.MISC + ITSIssueType.Misspelling
	require.Equal(t, rules.CatMisc.GetCategory(nil).GetID(), r.GetCategory().GetID())
	require.Equal(t, rules.ITSMisspelling, r.GetLocQualityIssueType())
	// Java 3-arg super → shortMessage null
	require.Equal(t, "", r.ShortDesc)
	require.Equal(t, "Dessa ord skrivs samman med bindestreck.", r.WithHyphenMessage)
	require.Equal(t, "Dessa ord skrivs samman.", r.WithoutHyphenMessage)
	require.Equal(t, "Dessa ord skrivs samman med eller utan bindestreck.", r.WithOrWithoutHyphenMessage)
	// Java AbstractCompoundRule default / field true
	require.True(t, r.SentenceStartsWithUpperCase)
	// Java getCompoundRuleData singleton from /sv/compounds.txt
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
			require.Equal(t, expSuggestions, matches[0].GetSuggestedReplacements(), "text %q", text)
		}
	}
	// correct:
	//check(0, "skit-bra"); // Java has this commented out (skit-bra+ → incorrect)
	check(0, "IP-Adress")
	check(0, "moll-tonart")
	check(0, "e-mail")
	// incorrect:
	check(1, "skit bra", "skitbra")
	check(1, "skit-bra", "skitbra")
	check(1, "IP Adress", "IP-Adress")
	check(1, "moll tonart", "moll-tonart", "molltonart")
	check(1, "e mail", "e-mail")
}
