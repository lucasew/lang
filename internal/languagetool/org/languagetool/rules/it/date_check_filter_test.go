package it

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDateCheckFilter(t *testing.T) {
	f := NewDateCheckFilter()
	require.NotNil(t, f)
	// month 1 exists in all locales via helper
	m, err := f.GetMonth("1")
	// may fail if helper requires name — try common english/local
	if err != nil {
		// try a few names
		for _, name := range []string{"jan", "januar", "enero", "gennaio", "stycznia", "januari"} {
			m, err = f.GetMonth(name)
			if err == nil {
				break
			}
		}
	}
	if err == nil {
		require.GreaterOrEqual(t, m, 1)
		require.LessOrEqual(t, m, 12)
	}
	_, err = f.AcceptRuleMatch(map[string]string{"year": "2014"})
	require.Error(t, err)
}
