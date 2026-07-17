package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildLintArgs(t *testing.T) {
	a := buildLintArgs("en", "text", "/data", "error", "", "", "", "", "", "", "", false, false, nil)
	require.Contains(t, a, "--lint")
	require.Contains(t, a, "-l")
	require.Contains(t, a, "en")
	require.Contains(t, a, "--data-dir")
	require.Contains(t, a, "-")

	a = buildLintArgs("en", "json", "", "warning", "", "", "X", "Y", "", "", "", true, true, []string{"f.txt", "g.txt"})
	require.Contains(t, a, "--format")
	require.Contains(t, a, "json")
	require.Contains(t, a, "--fail-on")
	require.Contains(t, a, "warning")
	require.Contains(t, a, "f.txt")
	require.Contains(t, a, "g.txt")
	require.Contains(t, a, "--enabledonly")
	require.Contains(t, a, "--recursive")
	require.Contains(t, a, "-e")
	require.Contains(t, a, "Y")
}
