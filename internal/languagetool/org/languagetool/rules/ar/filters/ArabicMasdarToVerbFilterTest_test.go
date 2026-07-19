package filters

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArabicMasdarToVerbFilter_Filter(t *testing.T) {
	f := NewArabicMasdarToVerbFilter()
	require.NotEmpty(t, f.Masdar2Verb, "official arabic_masdar_verb.txt should load")
	verbs := f.SuggestVerbsForMasdar("عمل")
	require.NotEmpty(t, verbs)
	sugs := f.SuggestionsFromArgs(map[string]string{"noun": "عمل"})
	require.NotEmpty(t, sugs)
	// Java inline invent map had إجابة; official file does not — fail closed
	require.Empty(t, f.SuggestVerbsForMasdar("إجابة"))
}
