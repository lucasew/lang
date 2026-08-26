package tools

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArabicConstantsAndPronouns(t *testing.T) {
	require.Equal(t, '\u0627', ArabicAlef)
	require.Equal(t, '\u0660', ArabicZero)
	require.Equal(t, "ني", GetAttachedPronoun("أنا"))
	require.Equal(t, "", GetAttachedPronoun("unknown"))
	require.Equal(t, "", GetAttachedPronoun(""))
}

func TestArabicUnitsHelper(t *testing.T) {
	require.True(t, IsArabicUnit("دينار"))
	require.False(t, IsArabicUnitFeminin("دينار"))
	require.True(t, IsArabicUnitFeminin("ليرة"))
	require.Equal(t, "ديناران", GetArabicUnitTwoForm("دينار", "raf3"))
	require.Equal(t, "ليراتٍ", GetArabicUnitPluralForm("ليرة", "nasb"))
	require.Equal(t, "[[unknown]]", GetArabicUnitOneForm("unknown", "raf3"))
}

func TestRemoveTashkeelStillWorks(t *testing.T) {
	// smoke: existing helper coexists with constants
	require.NotEmpty(t, TashkeelChars)
}
