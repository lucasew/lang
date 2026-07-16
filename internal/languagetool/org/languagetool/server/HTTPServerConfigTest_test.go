package server

// Twin of HTTPServerConfigTest
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of HTTPServerConfigTest.testArgumentParsing
func TestHTTPServerConfig_ArgumentParsing(t *testing.T) {
	c1, err := NewHTTPServerConfigFromArgs([]string{})
	require.NoError(t, err)
	require.Equal(t, DefaultPort, c1.Port)
	require.False(t, c1.PublicAccess)
	require.False(t, c1.Verbose)

	c2, err := NewHTTPServerConfigFromArgs([]string{"--public"})
	require.NoError(t, err)
	require.Equal(t, DefaultPort, c2.Port)
	require.True(t, c2.PublicAccess)
	require.False(t, c2.Verbose)

	c3, err := NewHTTPServerConfigFromArgs([]string{"--port", "80"})
	require.NoError(t, err)
	require.Equal(t, 80, c3.Port)
	require.False(t, c3.PublicAccess)

	c4, err := NewHTTPServerConfigFromArgs([]string{"--port", "80", "--public"})
	require.NoError(t, err)
	require.Equal(t, 80, c4.Port)
	require.True(t, c4.PublicAccess)
}

// Port of HTTPServerConfigTest.shouldLoadLanguageModelDirectoryFromCommandLineArguments
func TestHTTPServerConfig_ShouldLoadLanguageModelDirectoryFromCommandLineArguments(t *testing.T) {
	dir := t.TempDir()
	lm := filepath.Join(dir, "languageModelDirectory")
	require.NoError(t, os.MkdirAll(lm, 0o755))

	c, err := NewHTTPServerConfigFromArgs([]string{LanguageModelOption, lm})
	require.NoError(t, err)
	require.NotEmpty(t, c.LanguageModelDir)
	require.Equal(t, lm, c.LanguageModelDir)
	st, err := os.Stat(c.LanguageModelDir)
	require.NoError(t, err)
	require.True(t, st.IsDir())
	require.True(t, filepath.Base(c.LanguageModelDir) == "languageModelDirectory" ||
		filepath.Base(c.LanguageModelDir) == filepath.Base(lm))
}
