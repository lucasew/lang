package tagging

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeRussianSurface_Acute(t *testing.T) {
	// о + combining acute → strip to море
	in := "мо" + "\u0301" + "ре"
	got, yo := NormalizeRussianSurface(in)
	require.Equal(t, "море", got)
	// Java: contains("о́") → mayMissingYo stays false
	require.False(t, yo)
}

func TestNormalizeRussianSurface_MayMissingYo(t *testing.T) {
	got, yo := NormalizeRussianSurface("все")
	require.Equal(t, "все", got)
	require.True(t, yo) // е without ё / stress marks
}

func TestNormalizeRussianSurface_HardSign(t *testing.T) {
	got, _ := NormalizeRussianSurface("об\u02BCект")
	require.Equal(t, "объект", got)
}

func TestRussianMayMissingYoConfirmed(t *testing.T) {
	wt := MapWordTagger{
		"всё": {NewTaggedWord("всё", "ADV")},
	}
	// всё written as все
	n, flag := NormalizeRussianSurface("все")
	require.Equal(t, "все", n)
	require.True(t, flag)
	require.True(t, RussianMayMissingYoConfirmed(n, flag, wt))
	require.False(t, RussianMayMissingYoConfirmed("стол", true, wt))
}
