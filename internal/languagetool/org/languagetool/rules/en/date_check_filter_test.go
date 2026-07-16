package en

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDateCheckFilter(t *testing.T) {
	f := NewDateCheckFilter()
	d, err := f.GetDayOfWeekJava("Monday")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	m, err := f.GetMonth("January")
	require.NoError(t, err)
	require.Equal(t, 1, m)
	require.Equal(t, "Friday", f.GetDayOfWeekName(2014, 8, 29))
	_, err = f.AcceptRuleMatch(map[string]string{"year": "2014"})
	require.Error(t, err)
}
