package morfologik

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestMorfologikSpellerRule_EmojiIgnored(t *testing.T) {
	sp := NewMorfologikSpeller("/en/x.dict", 1)
	sp.AddWord("ok")
	r := NewMorfologikSpellerRule("MORFOLOGIK_RULE_EN_US", "en", "/en/x.dict", sp)
	r.IsMisspelled = sp.IsMisspelled
	sent := languagetool.AnalyzePlain("😂")
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, m)
}

func TestMorfologikSpellerRule_GetOnlySuggestions(t *testing.T) {
	sp := NewMorfologikSpeller("/en/x.dict", 1)
	sp.AddWord("ok")
	r := NewMorfologikSpellerRule("MORFOLOGIK_RULE_EN_US", "en", "/en/x.dict", sp)
	r.IsMisspelled = sp.IsMisspelled
	r.GetOnlySuggestionsFn = func(w string) []string {
		if w == "xyzzy" {
			return []string{"only"}
		}
		return nil
	}
	m, err := r.Match(languagetool.AnalyzePlain("xyzzy"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Equal(t, []string{"only"}, m[0].GetSuggestedReplacements())
}
