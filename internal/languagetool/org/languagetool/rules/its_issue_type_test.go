package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestITSIssueType(t *testing.T) {
	require.Equal(t, "grammar", ITSGrammar.String())
	require.Equal(t, "locale-specific-content", ITSLocaleSpecificContent.String())
	got, err := GetIssueType("misspelling")
	require.NoError(t, err)
	require.Equal(t, ITSMisspelling, got)
	_, err = GetIssueType("nope")
	require.Error(t, err)
	got, err = ParseIssueTypeCamel("LocaleSpecificContent")
	require.NoError(t, err)
	require.Equal(t, ITSLocaleSpecificContent, got)
}
