package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertCase(t *testing.T) {
	require.Equal(t, "", ConvertCase(CaseAllUpper, "", "X"))
	require.Equal(t, "hello", ConvertCase(CaseNone, "hello", "Hello"))
	require.Equal(t, "HELLO", ConvertCase(CaseAllUpper, "hello", "x"))
	require.Equal(t, "hello", ConvertCase(CaseAllLower, "HELLO", "x"))
	require.Equal(t, "Hello", ConvertCase(CaseStartUpper, "hello", "x"))
	require.Equal(t, "hELLO", ConvertCase(CaseStartLower, "HELLO", "x"))
	require.Equal(t, "Hello", ConvertCase(CaseFirstUpper, "hELLO", "x"))
	// preserve: sample all upper
	require.Equal(t, "HELLO", ConvertCase(CasePreserve, "hello", "WORLD"))
	// preserve: sample capitalized
	require.Equal(t, "Hello", ConvertCase(CasePreserve, "hello", "World"))
	// preserve: sample lower — leave token
	require.Equal(t, "hello", ConvertCase(CasePreserve, "hello", "world"))
}
