package ekavian

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikEkavianSpellerRule_Paths(t *testing.T) {
	r := NewMorfologikEkavianSpellerRule()
	require.Equal(t, MorfologikEkavianSpellerRuleID, r.GetID())
	require.Equal(t, EkavianSpellerDict, r.GetFileName())
	// Java getIgnoreFileName / getSpellingFileName / getProhibitFileName
	require.Equal(t, "/sr/dictionary/ekavian/ignored.txt", r.GetIgnoreFileName())
	require.Equal(t, "/sr/dictionary/ekavian/spelling.txt", r.GetSpellingFileName())
	require.Equal(t, "/sr/dictionary/ekavian/prohibit.txt", r.GetProhibitFileName())
}
