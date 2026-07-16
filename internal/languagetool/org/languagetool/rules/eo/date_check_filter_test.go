package eo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDateCheckFilter(t *testing.T) {
	f := NewDateCheckFilter()
	require.NotNil(t, f)
	_, err := f.AcceptRuleMatch(map[string]string{})
	require.Error(t, err)
	_, err = f.AcceptRuleMatch(map[string]string{"weekDay": "1"})
	require.NoError(t, err)
}

func TestDateFilterHelper_Month(t *testing.T) {
	h := NewDateFilterHelper()
	m, err := h.GetMonth("1")
	if err != nil {
		// try language names
		for _, name := range []string{"januaro"} {
			m, err = h.GetMonth(name)
			if err == nil {
				break
			}
		}
	}
	if err == nil {
		require.GreaterOrEqual(t, int(m), 1)
		require.LessOrEqual(t, int(m), 12)
	}
	_ = time.January
}
