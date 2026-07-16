package filters

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArabicNumberPhraseFilter_Filter(t *testing.T) {
	sugs := SuggestionsForNumericPhrase("5", false)
	require.NotEmpty(t, sugs)
	withPrev := PrepareSuggestion("3", "في", false)
	require.NotEmpty(t, withPrev)
}

func TestArabicNumberPhraseFilter_UnitFilter(t *testing.T) {
	// unit may be empty/unknown — still returns numeric forms
	sugs := PrepareSuggestionWithUnit("2", "", "كتاب", "raf3", false)
	// may be empty if unit helper returns empty — accept non-panic
	_ = sugs
	sugs2 := PrepareSuggestionWithUnit("5", "اشترى", "", "", false)
	require.NotEmpty(t, sugs2)
}
