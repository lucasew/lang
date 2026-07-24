package filters

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArabicVerbToMasdarFilter_Filter(t *testing.T) {
	f := NewArabicVerbToMasdarFilter()
	// default masdar "عمل" → verb forms include عمل variants
	masdars := f.SuggestMasdarsForVerb("عَمِلَ")
	require.NotEmpty(t, masdars)
}
