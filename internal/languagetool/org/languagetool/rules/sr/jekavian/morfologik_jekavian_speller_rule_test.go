package jekavian

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikJekavianSpellerRule_Paths(t *testing.T) {
	r := NewMorfologikJekavianSpellerRule()
	require.Equal(t, MorfologikJekavianSpellerRuleID, r.GetID())
	require.Equal(t, JekavianSpellerDict, r.GetFileName())
	require.Equal(t, "/sr/dictionary/jekavian/ignored.txt", r.GetIgnoreFileName())
	require.Equal(t, "/sr/dictionary/jekavian/spelling.txt", r.GetSpellingFileName())
	require.Equal(t, "/sr/dictionary/jekavian/prohibit.txt", r.GetProhibitFileName())
}
