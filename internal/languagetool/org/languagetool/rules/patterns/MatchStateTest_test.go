package patterns

// Twin of MatchStateTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of MatchStateTest.testConvertCase
func TestMatchState_ConvertCase(t *testing.T) {
	startUpper := NewMatchState(NewMatch("", "", false, "", "", CaseStartUpper, false, false, IncludeNone))
	require.Equal(t, "", startUpper.ConvertCase("", "Y", "en"))
	require.Equal(t, "X", startUpper.ConvertCase("x", "Y", "en"))
	require.Equal(t, "Xxx", startUpper.ConvertCase("xxx", "Yyy", "en"))
	require.Equal(t, "IJsselmeer", startUpper.ConvertCase("ijsselmeer", "Uppercase", "nl"))
	require.Equal(t, "IJ", startUpper.ConvertCase("ij", "Uppercase", "nl"))

	preserve := NewMatchState(NewMatch("", "", false, "", "", CasePreserve, false, false, IncludeNone))
	require.Equal(t, "Xxx", preserve.ConvertCase("xxx", "Yyy", "en"))
	require.Equal(t, "xxx", preserve.ConvertCase("xxx", "yyy", "en"))
	require.Equal(t, "XXX", preserve.ConvertCase("xxx", "YYY", "en"))
	require.Equal(t, "IJsselmeer", preserve.ConvertCase("ijsselmeer", "Uppercase", "nl"))
}
