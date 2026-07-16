package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContributor(t *testing.T) {
	c := NewContributorWithURL("Ada", "https://example.com")
	require.Equal(t, "Ada", c.GetName())
	require.Equal(t, "https://example.com", c.GetURL())
	require.Equal(t, "Daniel Naber", DanielNaber.GetName())
	err := NewRuleFilenameException("bad.xml")
	require.Contains(t, err.Error(), "bad.xml")
	require.Contains(t, err.Error(), "rules-en-English.xml")
}
