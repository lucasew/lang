package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSoftRuleMeta(t *testing.T) {
	id, name, issue, short := SoftRuleMeta("EN_A_VS_AN")
	require.Equal(t, "GRAMMAR", id)
	require.Equal(t, "Grammar", name)
	require.Equal(t, "grammar", issue)
	require.Equal(t, "Wrong article", short)

	id, _, issue, _ = SoftRuleMeta("EN_SOFT_YOUR_YOU_RE")
	require.Equal(t, "GRAMMAR", id)
	require.Equal(t, "grammar", issue)

	id, name, issue, short = SoftRuleMeta("EMPTY_LINE")
	require.Equal(t, "STYLE", id)
	require.Equal(t, "Style", name)
	require.Equal(t, "style", issue)
	require.Equal(t, "Empty line", short)
	require.Equal(t, "Empty line", SoftRuleDescription("EMPTY_LINE"))

	require.Equal(t, "error", SeverityFromIssueType("grammar"))
	require.Equal(t, "error", SeverityFromIssueType("misspelling"))
	require.Equal(t, "note", SeverityFromIssueType("style"))
	require.Equal(t, "warning", SeverityFromIssueType("whitespace"))

	require.Equal(t, "de", SoftRuleLangHint("DE_SOFT_DAS_DASS"))
	require.Equal(t, "fr", SoftRuleLangHint("FR_SOFT_A_LA"))
	require.Equal(t, "", SoftRuleLangHint("WHITESPACE_RULE"))
	require.Equal(t, "", SoftRuleLangHint("MORFOLOGIK_RULE_EN_US"))
	// empty lang → hint from rule id
	require.Contains(t, SoftRuleURL("DE_SOFT_DAS_DASS", ""), "lang=de")
	require.Contains(t, SoftRuleURL("EN_A_VS_AN", ""), "lang=en")
}
