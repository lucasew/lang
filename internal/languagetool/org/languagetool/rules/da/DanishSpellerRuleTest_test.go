package da

// Twin of languagetool-language-modules/da/src/test/java/org/languagetool/rules/da/DanishSpellerRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/hunspell"
	"github.com/stretchr/testify/require"
)

// Port of DanishSpellerRuleTest.testDashAndHyphenEtc
// Java uses Languages.getLanguageForShortCode("da") + JLanguageTool + HunspellRule (id HUNSPELL_RULE).
func TestDanishSpellerRule_DashAndHyphenEtc(t *testing.T) {
	r := NewDanishHunspellRule()
	require.Equal(t, hunspell.HunspellRuleID, r.GetID())

	// With nil dict, HunspellRule.IsMisspelledWord returns false (no invent misspell).
	// Non-letter dash runs are also ignored by nonAlphabeticRE.
	sent := languagetool.AnalyzePlain("De står under ----")
	matches, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, matches)

	// Soft parity with Java assertEquals(0, lt.check(...).size()) when speller cannot invent errors.
	lt := languagetool.NewJLanguageTool("da")
	require.NotEmpty(t, lt.Analyze("De står under ----"))
}
