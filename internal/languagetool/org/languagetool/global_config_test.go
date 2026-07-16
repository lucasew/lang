package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGlobalConfig(t *testing.T) {
	SetVerbose(true)
	require.True(t, IsVerbose())
	SetVerbose(false)
	c := &GlobalConfig{}
	c.SetGrammalecteServer("http://localhost")
	require.Equal(t, "http://localhost", c.GetGrammalecteServer())
	c2 := &GlobalConfig{GrammalecteServer: "http://localhost"}
	require.True(t, c.Equal(c2))
}
