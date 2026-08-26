package uk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpecialPOSTag_Hashtag(t *testing.T) {
	require.Equal(t, "hashtag", SpecialPOSTag("#янебоюсьсказати"))
}

func TestSpecialPOSTag_Numbers(t *testing.T) {
	require.Equal(t, "number", SpecialPOSTag("101,234"))
	require.Equal(t, "number", SpecialPOSTag("101 234"))
	require.Equal(t, "number", SpecialPOSTag("10–15"))
	require.Equal(t, "number:latin", SpecialPOSTag("XIX"))
	require.Equal(t, "number:latin", SpecialPOSTag("II"))
	require.Equal(t, "number:latin", SpecialPOSTag("X"))
	require.Equal(t, "date", SpecialPOSTag("14.07.2001"))
	require.Equal(t, "time", SpecialPOSTag("15:33"))
	require.Equal(t, "time", SpecialPOSTag("15:33:00"))
	require.Equal(t, "time", SpecialPOSTag("15.33"))
	require.Equal(t, "number:latin:bad:err", SpecialPOSTag("ХІХ"))
	require.Equal(t, "number:latin:bad", SpecialPOSTag("ІV"))
	require.Equal(t, "number:latin:bad", SpecialPOSTag("ХІХ-го"))
	require.Equal(t, "", SpecialPOSTag("D"))
}

func TestSpecialPOSTag_Degree(t *testing.T) {
	require.Equal(t, "number", SpecialPOSTag("7°"))
}
