package jekavian

// Twin of MorfologikJekavianSpellerRuleTest — map inject.
// Java isLatinScript() default true: pure Cyrillic ignored by pHasNoLetterLatin.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of MorfologikJekavianSpellerRuleTest.testMorfologikSpeller
func TestMorfologikJekavianSpellerRule_MorfologikSpeller(t *testing.T) {
	r := NewMorfologikJekavianSpellerRule()
	require.Equal(t, MorfologikJekavianSpellerRuleID, r.GetID())
	require.Equal(t, JekavianSpellerDict, r.GetFileName())

	sp := morfologik.NewMorfologikSpeller(JekavianSpellerDict, 1)
	for _, w := range []string{"zdravo", "svijet", "test"} {
		sp.AddWord(w)
	}
	sp.Suggestions["zdrav"] = []string{"zdravo"}
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled

	m, err := r.Match(languagetool.AnalyzePlain("zdravo svijet"))
	require.NoError(t, err)
	require.Empty(t, m)

	m, err = r.Match(languagetool.AnalyzePlain("zdrav"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetSuggestedReplacements(), "zdravo")
}

// Twin of MorfologikJekavianSpellerRuleTest.testSpellingCheck
func TestMorfologikJekavianSpellerRule_SpellingCheck(t *testing.T) {
	// alias of MorfologikSpeller path under audit name
	TestMorfologikJekavianSpellerRule_MorfologikSpeller(t)
}
