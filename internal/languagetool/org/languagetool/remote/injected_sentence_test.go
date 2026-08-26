package remote

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInjectedSentence(t *testing.T) {
	s := NewInjectedSentence("en", "  hello  ")
	require.Equal(t, "en", s.GetLanguage())
	require.Equal(t, "hello", s.GetText())
	require.True(t, s.Equal(NewInjectedSentence("en", "hello")))
	require.Contains(t, s.String(), "en")
}
