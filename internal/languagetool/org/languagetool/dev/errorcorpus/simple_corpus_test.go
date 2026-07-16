package errorcorpus

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSimpleCorpus_Parse(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ex.txt")
	require.NoError(t, os.WriteFile(path, []byte(
		"1. This is _a_ error. => an\n"+
			"2. Here _come_ another example. => comes\n"+
			"not a line\n",
	), 0o644))
	c, err := NewSimpleCorpus(path)
	require.NoError(t, err)
	require.Equal(t, 2, c.Len())
	require.True(t, c.HasNext())
	s1, err := c.Next()
	require.NoError(t, err)
	require.Equal(t, "This is a error.", s1.PlainText)
	require.Len(t, s1.Errors, 1)
	require.Equal(t, "an", s1.Errors[0].Correction)
	s2, err := c.Next()
	require.NoError(t, err)
	require.Equal(t, "comes", s2.Errors[0].Correction)
	require.False(t, c.HasNext())
}
