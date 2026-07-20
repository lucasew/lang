package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewYearDateFilterCore(t *testing.T) {
	jan := true
	y := 2014
	f := &NewYearDateFilterCore{ForceJanuary: &jan, ForceYear: &y}
	require.True(t, f.ShouldFlag(2013, 3))
	require.False(t, f.ShouldFlag(2013, 12))
	require.False(t, f.ShouldFlag(2014, 3))
	// Java always validates day with setLenient(false)
	msg := f.AcceptFromArgs(map[string]string{"year": "2013", "month": "5", "day": "10"}, "was {year} now {realYear}")
	require.Equal(t, "was 2013 now 2014", msg)
	// invalid day → suppress (not invent flag without day check)
	require.Empty(t, f.AcceptFromArgs(map[string]string{"year": "2013", "month": "2", "day": "30"}, "was {year}"))
	// soft hyphen stripped before day pattern (Java)
	require.Equal(t, "was 2013 now 2014", f.AcceptFromArgs(
		map[string]string{"year": "2013", "month": "5", "day": "1\u00AD0"}, "was {year} now {realYear}"))
}
