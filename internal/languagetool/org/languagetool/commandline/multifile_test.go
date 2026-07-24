package commandline

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseOptions_MultiFile(t *testing.T) {
	p := &CommandLineParser{}
	opts, err := p.ParseOptions([]string{"-l", "en", "--lint", "a.txt", "b.txt"})
	require.NoError(t, err)
	require.Equal(t, []string{"a.txt", "b.txt"}, opts.GetFilenames())
	require.Equal(t, "a.txt", opts.Filename)
}

func TestRunWithIO_MultiFileLint(t *testing.T) {
	dir := t.TempDir()
	f1 := filepath.Join(dir, "one.txt")
	f2 := filepath.Join(dir, "two.txt")
	require.NoError(t, os.WriteFile(f1, []byte("This is an test."), 0o644))
	require.NoError(t, os.WriteFile(f2, []byte("All good here."), 0o644))

	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--lint", f1, f2}, DefaultCoreHooks(), &out, &errb)
	require.Equal(t, 1, code, errb.String())
	body := out.String()
	require.Contains(t, body, "one.txt")
	require.Contains(t, body, "EN_A_VS_AN")
	// second file clean — still only one error overall
}

func TestParseOptions_OnlyAlias(t *testing.T) {
	p := &CommandLineParser{}
	opts, err := p.ParseOptions([]string{"--only", "-e", "EN_A_VS_AN", "-l", "en", "-"})
	require.NoError(t, err)
	require.True(t, opts.IsUseEnabledOnly())
	require.Equal(t, []string{"EN_A_VS_AN"}, opts.GetEnabledRules())
}
