package ekavian

// Twin of MorfologikEkavianSpellerRuleTest — map inject.
// Java isLatinScript() default true: pure Cyrillic has no Latin letters → ignoreWord
// (same as SpellingCheckRule.pHasNoLetterLatin). Misspell tests use Latin surfaces.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of MorfologikEkavianSpellerRuleTest.testMorfologikSpeller
func TestMorfologikEkavianSpellerRule_MorfologikSpeller(t *testing.T) {
	r := NewMorfologikEkavianSpellerRule()
	require.Equal(t, MorfologikEkavianSpellerRuleID, r.GetID())
	require.Equal(t, EkavianSpellerDict, r.GetFileName())

	sp := morfologik.NewMorfologikSpeller(EkavianSpellerDict, 1)
	for _, w := range []string{"zdravo", "svet", "test"} {
		sp.AddWord(w)
	}
	sp.Suggestions["zdrav"] = []string{"zdravo"}
	// Map-inject unit path: clear initSpeller Multis so Speller map is used.
	r.ClearMultiSpellers()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled

	m, err := r.Match(languagetool.AnalyzePlain("zdravo svet"))
	require.NoError(t, err)
	require.Empty(t, m)

	m, err = r.Match(languagetool.AnalyzePlain("zdrav"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetSuggestedReplacements(), "zdravo")

	// Pure Cyrillic: Java pHasNoLetterLatin → ignored under default isLatinScript
	m, err = r.Match(languagetool.AnalyzePlain("здраво"))
	require.NoError(t, err)
	require.Empty(t, m)
}
