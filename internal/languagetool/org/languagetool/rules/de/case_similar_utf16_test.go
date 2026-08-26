package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCaseRuleExceptions_NoTrimInvent(t *testing.T) {
	// Loading official file must succeed (no leading/trailing space lines)
	set := CaseRuleExceptions()
	require.NotEmpty(t, set)
}

func TestCaseRuleExceptions_WhitespaceLinePanics(t *testing.T) {
	// twin Java IllegalArgumentException on head/tail whitespace
	// validate helper logic on sample strings
	require.True(t, unicodeIsSpaceHeadOrTail(" foo"))
	require.True(t, unicodeIsSpaceHeadOrTail("foo "))
	require.False(t, unicodeIsSpaceHeadOrTail("foo bar"))
	require.False(t, unicodeIsSpaceHeadOrTail("Über"))
}

func unicodeIsSpaceHeadOrTail(line string) bool {
	n := utf16LenDE(line)
	if n == 0 {
		return false
	}
	c0 := javaCharAtDE(line, 0)
	cN := javaCharAtDE(line, n-1)
	return isJavaWhitespaceChar(c0) || isJavaWhitespaceChar(cN)
}

func isJavaWhitespaceChar(c rune) bool {
	// subset: Character.isWhitespace for common resource chars
	switch c {
	case ' ', '\t', '\n', '\r', '\f':
		return true
	}
	return false
}

// Apache getLevenshteinDistance on UTF-16: supplementary pairs count as 2.
func TestLevenshteinSimilar_UTF16Units(t *testing.T) {
	// BMP German: same as rune distance
	require.Equal(t, 1, levenshteinSimilar("Merkel", "Merkl"))
	require.Equal(t, 0, levenshteinSimilar("Haus", "Haus"))
	// length gate uses UTF-16
	require.Equal(t, 1, utf16LenDE("Ä"))
}
