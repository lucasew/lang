package en

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnglishPartialPosTagFilter(t *testing.T) {
	f := NewNoDisambiguationEnglishPartialPosTagFilter(func(p string) []string {
		if p == "happy" {
			return []string{"JJ"}
		}
		return nil
	})
	ok, err := f.Accept("unhappy", "un(.*)", "JJ", false, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)
}
