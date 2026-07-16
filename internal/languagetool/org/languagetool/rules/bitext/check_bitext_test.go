package bitext

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of ToolsTest.testBitextCheck
func TestCheckBitext(t *testing.T) {
	matches := CheckBitext(
		"This is a perfectly good sentence.",
		"To jest całkowicie prawidłowe zdanie.",
		nil,
	)
	require.NotNil(t, matches)

	same := CheckBitext("Hello world.", "Hello world.", nil)
	require.NotEmpty(t, same, "expected bitext matches for identical src/trg")
}
