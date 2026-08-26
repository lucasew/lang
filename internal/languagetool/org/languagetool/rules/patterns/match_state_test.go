package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatchStateConvertCasePort(t *testing.T) {
	startUpper := NewMatchState(NewMatch("", "", false, "", "", CaseStartUpper, false, false, IncludeNone))
	require.Nil(t, nil) // placeholder for null string — ConvertCase returns empty for empty
	require.Equal(t, "", startUpper.ConvertCase("", "Y", "en"))
	require.Equal(t, "X", startUpper.ConvertCase("x", "Y", "en"))
	require.Equal(t, "Xxx", startUpper.ConvertCase("xxx", "Yyy", "en"))
	// Dutch IJ
	require.Equal(t, "IJsselmeer", startUpper.ConvertCase("ijsselmeer", "Uppercase", "nl"))
	require.Equal(t, "IJ", startUpper.ConvertCase("ij", "Uppercase", "nl"))

	preserve := NewMatchState(NewMatch("", "", false, "", "", CasePreserve, false, false, IncludeNone))
	require.Equal(t, "Xxx", preserve.ConvertCase("xxx", "Yyy", "en"))
	require.Equal(t, "xxx", preserve.ConvertCase("xxx", "yyy", "en"))
	require.Equal(t, "XXX", preserve.ConvertCase("xxx", "YYY", "en"))
	require.Equal(t, "IJsselmeer", preserve.ConvertCase("ijsselmeer", "Uppercase", "nl"))
	require.Equal(t, "ijsselmeer", preserve.ConvertCase("ijsselmeer", "lowercase", "nl"))
	require.Equal(t, "IJSSELMEER", preserve.ConvertCase("ijsselmeer", "ALLUPPER", "nl"))

	startLower := NewMatchState(NewMatch("", "", false, "", "", CaseStartLower, false, false, IncludeNone))
	require.Equal(t, "xxx", startLower.ConvertCase("xxx", "YYY", "en"))
	require.Equal(t, "xXX", startLower.ConvertCase("XXX", "Yyy", "en"))
	require.Equal(t, "xxx", startLower.ConvertCase("Xxx", "Yyy", "en"))
}
