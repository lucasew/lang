package tools

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNumberToArabicWords(t *testing.T) {
	// Twin of ArabicNumbersWords.numberToArabicWords core cases.
	require.Equal(t, "صفر", NumberToArabicWords("0"))
	require.Equal(t, "واحد", NumberToArabicWords("1"))
	require.Equal(t, "اثنان", NumberToArabicWords("2"))
	require.Equal(t, "ثلاثة", NumberToArabicWords("3"))
	require.Equal(t, "عشرة", NumberToArabicWords("10"))
	require.Equal(t, "أحد عشر", NumberToArabicWords("11"))
	require.Equal(t, "واحد وعشرون", NumberToArabicWords("21"))
	require.Equal(t, "مائة", NumberToArabicWords("100"))
	require.Equal(t, "مائة وخمسة", NumberToArabicWords("105"))
	require.Equal(t, "ألف", NumberToArabicWords("1000"))
	require.Equal(t, "ألفان", NumberToArabicWords("2000"))
	// Java early-return for ONE ignores feminine flag.
	require.Equal(t, "واحد", NumberToArabicWordsGender("1", true))
	// jar inflection for 2
	require.Equal(t, "اثنين", NumberToArabicWordsInflected("2", false, false, "jar"))
}

func TestTextToNumber(t *testing.T) {
	require.Equal(t, 0, TextToNumber("صفر"))
	require.Equal(t, 1, TextToNumber("واحد"))
	require.Equal(t, 2, TextToNumber("اثنان"))
	require.Equal(t, 11, TextToNumber("أحد عشر"))
	require.Equal(t, 25, TextToNumber("خمسة وعشرون"))
	require.Equal(t, 100, TextToNumber("مائة"))
	require.Equal(t, 1000, TextToNumber("ألف"))
	require.Equal(t, 10000, TextToNumber("عشرة آلاف"))
}

func TestGetSuggestionsNumericPhrase(t *testing.T) {
	// jar form of 2 should correct "اثنان" when inflection=jar
	sugs := GetSuggestionsNumericPhrase("اثنان", false, false, "jar")
	require.Equal(t, []string{"اثنين"}, sugs)
	// correct jar form → no suggestion
	sugs = GetSuggestionsNumericPhrase("اثنين", false, false, "jar")
	require.Empty(t, sugs)
	// empty inflection → both raf3 and jar when incorrect
	sugs = GetSuggestionsNumericPhrase("اثنين", false, false, "")
	require.Len(t, sugs, 2)
	require.Equal(t, "اثنان", sugs[0])
	require.Equal(t, "اثنين", sugs[1])
}
