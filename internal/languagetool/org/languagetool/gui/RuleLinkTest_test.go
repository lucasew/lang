package gui

// Twin of RuleLinkTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuleLink_BuildDeactivationLink(t *testing.T) {
	link := BuildDeactivationLink("WORD_REPEAT_RULE")
	require.Equal(t, "WORD_REPEAT_RULE", link.GetID())
	require.Equal(t, "http://languagetool.org/deactivate/WORD_REPEAT_RULE", link.String())
}

func TestRuleLink_BuildReactivationLink(t *testing.T) {
	link := BuildReactivationLink("WORD_REPEAT_RULE")
	require.Equal(t, "WORD_REPEAT_RULE", link.GetID())
	require.Equal(t, "http://languagetool.org/reactivate/WORD_REPEAT_RULE", link.String())
}

func TestRuleLink_GetFromString(t *testing.T) {
	ruleLink1, err := GetRuleLinkFromString("http://languagetool.org/reactivate/FOO_BAR_ID")
	require.NoError(t, err)
	require.Equal(t, "FOO_BAR_ID", ruleLink1.GetID())
	require.Equal(t, "http://languagetool.org/reactivate/FOO_BAR_ID", ruleLink1.String())

	ruleLink2, err := GetRuleLinkFromString("http://languagetool.org/deactivate/FOO_BAR_ID2")
	require.NoError(t, err)
	require.Equal(t, "FOO_BAR_ID2", ruleLink2.GetID())
	require.Equal(t, "http://languagetool.org/deactivate/FOO_BAR_ID2", ruleLink2.String())
}
