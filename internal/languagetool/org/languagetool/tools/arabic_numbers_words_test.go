package tools

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNumberToArabicWords(t *testing.T) {
	require.Equal(t, "صفر", NumberToArabicWords("0"))
	require.Equal(t, "واحد", NumberToArabicWords("1"))
	require.Equal(t, "اثنان", NumberToArabicWords("2"))
	require.Equal(t, "ثلاثة", NumberToArabicWords("3"))
	require.Equal(t, "عشرة", NumberToArabicWords("10"))
	require.Equal(t, "أحد عشر", NumberToArabicWords("11"))
	require.Contains(t, NumberToArabicWords("21"), "واحد")
	require.Contains(t, NumberToArabicWords("21"), "عشرون")
	require.Equal(t, "مائة", NumberToArabicWords("100"))
	require.Contains(t, NumberToArabicWords("105"), "مائة")
	require.Contains(t, NumberToArabicWords("1000"), "ألف")
	require.Contains(t, NumberToArabicWords("2000"), "ألفان")
	require.Equal(t, "إحدى", NumberToArabicWordsGender("1", true))
}
