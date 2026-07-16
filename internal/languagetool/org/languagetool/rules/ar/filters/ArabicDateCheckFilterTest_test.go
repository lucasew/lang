package filters

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArabicDateCheckFilter_AcceptIncompleteArgs(t *testing.T) {
	f := NewArabicDateCheckFilter()
	// ValidateDateFilterArgs rejects incomplete maps
	require.Error(t, ValidateDateFilterArgs(map[string]string{}))
	require.NoError(t, ValidateDateFilterArgs(map[string]string{"weekDay": "الأحد"}))
	// known month/day helpers
	m, err := f.GetMonth("يناير")
	require.NoError(t, err)
	require.Equal(t, 1, m)
}
