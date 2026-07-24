package ca

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCatalanRemoteRewriteConfig(t *testing.T) {
	c := DefaultRemoteRewriteConfig()
	// typically no env in tests
	require.False(t, c.IsRemoteServiceAvailable() && c.ServerURL == "")
	c.ServerURL = "http://localhost"
	require.True(t, c.IsRemoteServiceAvailable())
	require.True(t, c.AcceptsSentence("Hola"))
	require.False(t, c.AcceptsSentence(""))
	c.MaxChars = 5
	require.False(t, c.AcceptsSentence("123456"))
	_ = strings.TrimSpace
}
