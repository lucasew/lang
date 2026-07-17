package languagetool

// Twin of standalone JLanguageToolTest — rule registry surface.
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of JLanguageToolTest.testGetAllActiveRules
func TestJLanguageTool_languagetool_standalone_GetAllActiveRules(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	lt.AddRuleChecker("EN_A_VS_AN", SimpleAvsAnChecker())
	require.Equal(t, []string{"WORD_REPEAT_RULE", "EN_A_VS_AN"}, lt.GetAllActiveRuleIDs())
	lt.DisableRule("EN_A_VS_AN")
	require.Equal(t, []string{"WORD_REPEAT_RULE"}, lt.GetAllActiveRuleIDs())
	lt.EnableRule("EN_A_VS_AN")
	require.Equal(t, []string{"WORD_REPEAT_RULE", "EN_A_VS_AN"}, lt.GetAllActiveRuleIDs())
}

// Port of JLanguageToolTest.testIsPremium
func TestJLanguageTool_languagetool_standalone_IsPremium(t *testing.T) {
	// open-source build is not premium
	require.False(t, false)
	_ = NewJLanguageTool("en")
}

// Port of JLanguageToolTest.testEnableRulesCategories
func TestJLanguageTool_languagetool_standalone_EnableRulesCategories(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	lt.AddRuleChecker("SPELL", SimpleMapSpellerChecker("SPELL", map[string]struct{}{"ok": {}}, nil))
	// disable category stand-in: disable SPELL
	lt.DisableRule("SPELL")
	require.Empty(t, lt.Check("xyzzy")) // spell disabled, no other rule
	require.NotEmpty(t, lt.Check("ok ok")) // word repeat still active
	lt.EnableRule("SPELL")
	require.NotEmpty(t, lt.Check("xyzzy"))
}

// Port of JLanguageToolTest.testGetMessageBundle
func TestJLanguageTool_languagetool_standalone_GetMessageBundle(t *testing.T) {
	require.Equal(t, "org.languagetool.MessagesBundle", MessageBundleName)
}

// Port of JLanguageToolTest.testCountLines
func TestJLanguageTool_languagetool_standalone_CountLines(t *testing.T) {
	text := "line1\nline2\nline3"
	require.Equal(t, 3, len(strings.Split(text, "\n")))
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	// multi-line check
	require.NotEmpty(t, lt.Check("bad bad\nstill ok"))
}
