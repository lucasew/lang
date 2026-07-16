package filters

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArabicMasdarToVerbFilter_Filter(t *testing.T) {
	f := NewArabicMasdarToVerbFilter()
	verbs := f.SuggestVerbsForMasdar("عمل")
	require.NotEmpty(t, verbs)
	sugs := f.SuggestionsFromArgs(map[string]string{"noun": "عمل"})
	require.NotEmpty(t, sugs)
}
