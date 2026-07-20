package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContributor(t *testing.T) {
	c := NewContributorWithURL("Ada", "https://example.com")
	require.Equal(t, "Ada", c.GetName())
	require.Equal(t, "https://example.com", c.GetURL())
	require.Equal(t, "Ada", c.String())
	// empty name allowed (Java requireNonNull only rejects null)
	require.Equal(t, "", NewContributor("").GetName())
	require.Equal(t, "", NewContributor("x").GetURL())
	require.Equal(t, "Daniel Naber", DanielNaber.GetName())
	require.Equal(t, "http://www.danielnaber.de", DanielNaber.GetURL())
	err := NewRuleFilenameException("bad.xml")
	require.Contains(t, err.Error(), "bad.xml")
	require.Contains(t, err.Error(), "rules-en-English.xml")
}
