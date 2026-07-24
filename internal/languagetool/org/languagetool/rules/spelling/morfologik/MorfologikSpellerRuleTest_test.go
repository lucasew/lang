package morfologik

// Twin of languagetool-core MorfologikSpellerRuleTest (Java class has no @Test methods).
// Surface inject smoke for rule Match / ID / file name.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of MorfologikSpellerRuleTest (no @Test) — green inject path.
func TestMorfologikSpellerRule_NoTests(t *testing.T) {
	sp := NewMorfologikSpeller("/xx/spelling/test.dict", 1)
	sp.AddWord("hello")
	sp.Suggestions["helo"] = []string{"hello"}
	r := NewMorfologikSpellerRule("MORFOLOGIK_RULE_XX", "xx", sp.FileInClassPath, sp)
	require.Equal(t, "MORFOLOGIK_RULE_XX", r.GetID())
	require.Equal(t, "/xx/spelling/test.dict", r.GetFileName())
	m, err := r.Match(languagetool.AnalyzePlain("hello helo"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Equal(t, []string{"hello"}, m[0].GetSuggestedReplacements())
}
