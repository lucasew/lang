package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGlobalConfig(t *testing.T) {
	SetVerbose(true)
	require.True(t, IsVerbose())
	SetVerbose(false)
	require.False(t, IsVerbose())
	c := &GlobalConfig{}
	c.SetGrammalecteServer("http://localhost")
	require.Equal(t, "http://localhost", c.GetGrammalecteServer())
	c2 := &GlobalConfig{GrammalecteServer: "http://localhost"}
	require.True(t, c.Equal(c2))
	// hash ignores user/password (Java hashCode)
	c.SetGrammalecteUser("u")
	c3 := &GlobalConfig{GrammalecteServer: "http://localhost"}
	require.Equal(t, c.Hash(), c3.Hash())
	require.False(t, c.Equal(c3))
}
