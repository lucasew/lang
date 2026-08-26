package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildLintArgs(t *testing.T) {
	a := buildLintArgs(lintArgs{lang: "en", format: "text", dataDir: "/data", failOn: "error"})
	require.Contains(t, a, "--lint")
	require.Contains(t, a, "-l")
	require.Contains(t, a, "en")
	require.Contains(t, a, "--data-dir")
	require.Contains(t, a, "-")

	a = buildLintArgs(lintArgs{
		lang: "en", format: "json", failOn: "warning", disable: "X", enable: "Y",
		enabledOnly: true, recursive: true, ignoreWords: "xyzzy,foo", files: []string{"f.txt", "g.txt"},
	})
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
	require.Contains(t, a, "--ignore-words")
	require.Contains(t, a, "xyzzy,foo")

	a = buildLintArgs(lintArgs{lang: "en", apply: true, mother: "de", files: []string{"-"}})
	require.Contains(t, a, "--apply")
	require.Contains(t, a, "-m")
	require.Contains(t, a, "de")
	require.NotContains(t, a, "--lint")

	a = buildLintArgs(lintArgs{
		lang: "en", ignoreSpellingFile: "/tmp/ign.txt", disambiguationFile: "/tmp/d.xml",
	})
	require.Contains(t, a, "--ignore-spelling-file")
	require.Contains(t, a, "/tmp/ign.txt")
	require.Contains(t, a, "--disambiguation-file")
	require.Contains(t, a, "/tmp/d.xml")
}
