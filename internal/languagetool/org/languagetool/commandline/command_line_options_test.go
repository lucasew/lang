package commandline

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommandLineOptions_JavaSurface(t *testing.T) {
	o := NewCommandLineOptions()
	require.Equal(t, OutputPlaintext, o.GetOutputFormat())
	require.Equal(t, "DEFAULT", o.GetLevel())
	o.SetJsonFormat()
	require.True(t, o.IsJsonFormat())
	o.SetRemoteRulesFile("/remote.json")
	require.Equal(t, "/remote.json", o.GetRemoteRulesFile())
	o.SetBitextRuleFile("/bitext.xml")
	require.Equal(t, "/bitext.xml", o.GetBitextRuleFile())
	o.SetUseEnabledOnly()
	require.True(t, o.IsUseEnabledOnly())
	o.SetEnabledCategories([]string{"STYLE"})
	require.Equal(t, []string{"STYLE"}, o.GetEnabledCategories())
}
