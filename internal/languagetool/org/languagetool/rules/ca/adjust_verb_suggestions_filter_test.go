package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdjustVerbSuggestionsFilter(t *testing.T) {
	f := NewAdjustVerbSuggestionsFilter()
	got := f.Suggest(VerbSuggestionContext{
		PronounsStr:   "em",
		VerbStr:       "vaig",
		WholeOriginal: "em vaig",
	}, "replaceEmEn")
	require.Equal(t, []string{"en vaig"}, got)

	got = f.Suggest(VerbSuggestionContext{
		PronounsStr:            "",
		VerbStr:                "menja",
		FirstVerbPersonaNumber: "3S",
		WholeOriginal:          "menja",
	}, "addPronounReflexive")
	require.NotEmpty(t, got)
}
