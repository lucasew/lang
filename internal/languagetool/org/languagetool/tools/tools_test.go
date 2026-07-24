package tools

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestI18n(t *testing.T) {
	require.Equal(t, "Hello world", I18n("Hello {0}", "world"))
	require.Equal(t, "a and b", CorrectListToString([]string{"a", "b"}, "and"))
	require.Equal(t, "a, b, and c", CorrectListToString([]string{"a", "b", "c"}, "and"))
	// MessageFormat: '' → ', and '…' is literal (no {0} expand)
	require.Equal(t, "It's {0}", I18n("It''s '{0}'", "X"))
	require.Equal(t, "x=val%ue", I18n("x={0}", "val%ue")) // no fmt verbs on args
	require.Equal(t, "a-b", I18n("{0}-{1}", "a", "b"))
}

func TestGetFullStackTrace(t *testing.T) {
	require.Equal(t, "", GetFullStackTrace(nil))
	s := GetFullStackTrace(errors.New("boom"))
	require.Contains(t, s, "boom")
	require.Contains(t, s, "goroutine")
}
