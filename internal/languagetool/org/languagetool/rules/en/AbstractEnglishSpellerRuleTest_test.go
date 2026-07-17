package en

// Twin of AbstractEnglishSpellerRuleTest (Java has no @Test) — surface smoke.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of AbstractEnglishSpellerRuleTest (no @Test)
func TestAbstractEnglishSpellerRule_NoTests(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller("/en/hunspell/en_US.dict", 1)
	sp.AddWord("colour")
	sp.AddWord("color")
	sp.Suggestions["collor"] = []string{"color", "colour"}
	r := NewAbstractEnglishSpellerRule("MORFOLOGIK_RULE_EN_US", "en-US", sp)
	require.Equal(t, "MORFOLOGIK_RULE_EN_US", r.GetID())
	require.Equal(t, "en", r.LanguageShortCode)
	require.Equal(t, "en-US", r.VariantCode)
	require.Contains(t, r.GetAdditionalSpellingFileNames(), "en/hunspell/spelling.txt")
	require.True(t, IsDoNotSuggest("bullshit"))
	require.False(t, IsDoNotSuggest("hello"))
	require.Equal(t, []string{"color"}, FilterEnglishSuggestions([]string{"color", "bullshit"}))

	m, err := r.Match(languagetool.AnalyzePlain("color collor"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetSuggestedReplacements(), "color")
}
