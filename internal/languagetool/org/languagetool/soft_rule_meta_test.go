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

	require.Equal(t, "error", SeverityFromIssueType("grammar"))
	require.Equal(t, "error", SeverityFromIssueType("misspelling"))
	require.Equal(t, "note", SeverityFromIssueType("style"))
	require.Equal(t, "warning", SeverityFromIssueType("whitespace"))
}
