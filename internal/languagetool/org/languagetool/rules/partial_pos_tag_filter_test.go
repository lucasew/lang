package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPartialPosTagFilter(t *testing.T) {
	f := NewPartialPosTagFilter(func(partial string) []string {
		if partial == "happy" {
			return []string{"JJ"}
		}
		return []string{"NN"}
	})
	ok, err := f.Accept("unhappy", "un(.*)", "JJ", false, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = f.Accept("unhappy", "un(.*)", "VB", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)

	ok, err = f.Accept("unhappy", "un(.*)", "JJ", true, false, "", "")
	require.NoError(t, err)
	require.False(t, ok) // negated: has JJ → false
}
