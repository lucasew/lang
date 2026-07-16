package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnalyzedTokenReadings_IgnoreSpelling(t *testing.T) {
	r := NewAnalyzedTokenReadings(NewAnalyzedToken("foo", nil, nil))
	require.False(t, r.IsIgnoredBySpeller())
	r.IgnoreSpelling()
	require.True(t, r.IsIgnoredBySpeller())

	// FromOld copies flag
	copy := NewAnalyzedTokenReadingsFromOld(r, []*AnalyzedToken{NewAnalyzedToken("foo", nil, nil)}, "rule")
	require.True(t, copy.IsIgnoredBySpeller())
}
