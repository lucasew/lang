package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdjustPronounsFilter(t *testing.T) {
	f := NewAdjustPronounsFilter()
	ctx := PronounVerbContext{
		PronounsStr:   "em",
		VerbStr:       "vaig",
		WholeOriginal: "em vaig",
		CasingModel:   "em",
	}
	got := f.Suggest(ctx, "replaceEmEn")
	require.Equal(t, []string{"en vaig"}, got)

	ctx2 := PronounVerbContext{
		PronounsStr:            "ho",
		VerbStr:                "menja",
		FirstVerbPersonaNumber: "3S",
		PronounsAfter:          false,
		WholeOriginal:          "ho menja",
	}
	got = f.Suggest(ctx2, "addPronounReflexive")
	require.NotEmpty(t, got)
	require.Contains(t, got[0], "menja")
}
