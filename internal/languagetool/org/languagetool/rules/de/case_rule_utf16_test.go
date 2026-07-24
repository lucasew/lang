package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

// Twin: CaseRule charAt(0) / length() use UTF-16 (via tools.StartsWith* and utf16LenDE).
func TestCaseRule_CharAtUTF16Helpers(t *testing.T) {
	require.True(t, tools.StartsWithLowercase("essen"))
	require.True(t, tools.StartsWithUppercase("Essen"))
	require.False(t, tools.StartsWithLowercase(""))
	// abbreviation length gate: single UTF-16 unit skipped
	require.False(t, utf16LenDE("A") > 1)
	require.True(t, utf16LenDE("Ab") > 1)
	// invisible separator U+2063
	tok := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("\u2063", nil, nil),
	}, 0)
	require.True(t, isInvisibleSeparator(0, []*languagetool.AnalyzedTokenReadings{tok}))
	require.False(t, isInvisibleSeparator(0, []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
			languagetool.NewAnalyzedToken("x", nil, nil),
		}, 0),
	}))
}

func TestJavaCharAtDE(t *testing.T) {
	require.Equal(t, 'E', javaCharAtDE("Essen", 0))
	require.Equal(t, rune(0), javaCharAtDE("", 0))
	require.Equal(t, '\u2063', javaCharAtDE("\u2063", 0))
}
