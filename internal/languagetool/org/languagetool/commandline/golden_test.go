package commandline

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCoreGoldenAndCompare_CLI(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"golden", "-l", "en", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "This is an test.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1, errb.String()+out.String())
	var findings []Finding
	require.NoError(t, json.Unmarshal(out.Bytes(), &findings))
	require.NotEmpty(t, findings)

	tmp := filepath.Join(t.TempDir(), "g.json")
	require.NoError(t, os.WriteFile(tmp, out.Bytes(), 0o644))

	out.Reset()
	errb.Reset()
	code = RunWithIO([]string{"compare", tmp, "-l", "en", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "This is an test.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.Equal(t, 0, code, errb.String()+out.String())
	require.Contains(t, out.String(), "OK")
}

func TestCoreGoldenHook_Direct(t *testing.T) {
	var buf bytes.Buffer
	n, err := CoreGoldenHook(&buf, "This is an test.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	require.GreaterOrEqual(t, n, 1)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_A_VS_AN" {
			found = true
			require.Equal(t, "error", f.Severity)
			require.Equal(t, "grammar", f.Type)
			require.Equal(t, "a", f.Suggestion)
		}
	}
	require.True(t, found)

	// compare against itself
	tmp := filepath.Join(t.TempDir(), "golden.json")
	require.NoError(t, os.WriteFile(tmp, buf.Bytes(), 0o644))
	var cout bytes.Buffer
	n, err = CoreCompareHook(&cout, "This is an test.", &CommandLineOptions{
		Language: "en", CompareGoldenPath: tmp,
	})
	require.NoError(t, err)
	require.Equal(t, 0, n)
	require.Contains(t, cout.String(), "OK")
}

func TestCompareFindings_Diff(t *testing.T) {
	got := []Finding{{Rule: "A", Message: "m", Offset: 1, Length: 1, Type: "grammar", Severity: "error"}}
	want := []Finding{{Rule: "A", Message: "m", Offset: 1, Length: 1, Type: "style", Severity: "note"}}
	diff := CompareFindings(got, want)
	require.Contains(t, diff, "type")
}

func TestNormalizeProductArgs_GoldenCompare(t *testing.T) {
	require.Equal(t, []string{"--golden", "-l", "en"}, NormalizeProductArgs([]string{"golden", "-l", "en"}))
	require.Equal(t, []string{"--compare", "g.json", "-l", "en"}, NormalizeProductArgs([]string{"compare", "g.json", "-l", "en"}))
	_ = runtime.Version()
}
