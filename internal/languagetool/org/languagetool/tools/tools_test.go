package tools

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestI18n(t *testing.T) {
	require.Equal(t, "Hello world", I18n("Hello {0}", "world"))
	require.Equal(t, "a and b", CorrectListToString([]string{"a", "b"}, "and"))
	require.Equal(t, "a, b, and c", CorrectListToString([]string{"a", "b", "c"}, "and"))
}
