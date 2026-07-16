package morfologik

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestMorfologikSpellerAndRule(t *testing.T) {
	sp := NewMorfologikSpeller("/en/spelling.dict", 1)
	sp.AddWord("cat")
	sp.AddWord("dog")
	sp.Suggestions["cta"] = []string{"cat"}
	require.True(t, sp.IsMisspelled("cta"))
	require.False(t, sp.IsMisspelled("cat"))
	require.Equal(t, []string{"cat"}, sp.FindReplacements("cta"))

	r := NewMorfologikSpellerRule("MORFOLOGIK_RULE_EN_US", "en-US", sp.FileInClassPath, sp)
	sent := languagetool.AnalyzePlain("cat cta dog")
	matches, err := r.Match(sent)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	require.Equal(t, "cta", sent.GetText()[matches[0].FromPos:matches[0].ToPos])

	multi := NewMorfologikMultiSpeller(sp, NewMorfologikSpeller("user", 1))
	require.False(t, multi.IsMisspelled("cat"))
	require.True(t, multi.IsMisspelled("zzz"))
}
