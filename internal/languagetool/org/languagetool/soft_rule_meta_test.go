package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSoftRuleMeta_KnownJavaFamilies(t *testing.T) {
	id, name, issue, short := SoftRuleMeta("EN_A_VS_AN")
	require.Equal(t, "GRAMMAR", id)
	require.Equal(t, "Grammar", name)
	require.Equal(t, "grammar", issue)
	require.NotEmpty(t, short)

	id, _, issue, _ = SoftRuleMeta("MORFOLOGIK_RULE_EN_US")
	require.Equal(t, "TYPOS", id)
	require.Equal(t, "misspelling", issue)

	id, _, issue, _ = SoftRuleMeta("WHITESPACE_RULE")
	require.Equal(t, "TYPOGRAPHY", id)
	require.Equal(t, "whitespace", issue)

	// Soft invent IDs must not get special grammar/style invent — uncategorized.
	id, _, issue, _ = SoftRuleMeta("EN_SOFT_YOUR_YOU_RE")
	require.Equal(t, "MISC", id)
	require.Equal(t, "uncategorized", issue)

	id, name, issue, short = SoftRuleMeta("EMPTY_LINE")
	require.Equal(t, "STYLE", id)
	require.Equal(t, "style", issue)
	require.Equal(t, "Empty line", short)

	require.Equal(t, "error", SeverityFromIssueType("grammar"))
	require.Equal(t, "error", SeverityFromIssueType("misspelling"))
	require.Equal(t, "note", SeverityFromIssueType("style"))
	require.Equal(t, "warning", SeverityFromIssueType("whitespace"))
}

func TestSoftRuleDescription_Known(t *testing.T) {
	require.Equal(t, "Use of 'a' versus 'an'", SoftRuleDescription("EN_A_VS_AN"))
	require.Equal(t, "Empty line", SoftRuleDescription("EMPTY_LINE"))
	// Soft invent: description is the raw id, not a fancy invent label.
	require.Equal(t, "EN_SOFT_YOUR_YOU_RE", SoftRuleDescription("EN_SOFT_YOUR_YOU_RE"))
}

func TestSoftRuleLangHint(t *testing.T) {
	require.Equal(t, "de", SoftRuleLangHint("DE_AGREEMENT"))
	require.Equal(t, "fr", SoftRuleLangHint("FR_AGREEMENT"))
	require.Equal(t, "", SoftRuleLangHint("UNKNOWN"))
}

func TestSoftRuleURL(t *testing.T) {
	require.Contains(t, SoftRuleURL("EN_A_VS_AN", "en"), "lang=en")
	require.Contains(t, SoftRuleURL("DE_AGREEMENT", ""), "lang=de")
}
