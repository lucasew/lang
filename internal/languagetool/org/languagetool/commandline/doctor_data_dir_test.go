package commandline

import (
	"bytes"
	"io"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCoreDoctor(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, CoreDoctor(&buf, nil))
	out := buf.String()
	require.Contains(t, out, "lang doctor")
	require.Contains(t, out, "EN_A_VS_AN ok")
	require.Contains(t, out, "status: ok")
	require.Contains(t, out, "corepack languages:")
	// multiwords path or embedded soft defaults
	require.True(t, strings.Contains(out, "en multiwords:"), out)
	require.Contains(t, out, "en soft rules:")
	require.Contains(t, out, "en soft smoke:")
	// walk-up usually finds testdata/grammar with many *-soft.xml packs
	if strings.Contains(out, "grammar dir: (unset)") {
		t.Log("grammar dir unset in this environment")
	} else {
		require.Contains(t, out, "soft grammar files:")
	}
}

func TestRunWithIO_Doctor(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"doctor"}, DefaultCoreHooks(), &out, &errb)
	require.Equal(t, 0, code, errb.String())
	require.Contains(t, out.String(), "status: ok")
}

func TestResolveGrammarDir_DataDir(t *testing.T) {
	opts := NewCommandLineOptions()
	opts.SetDataDir("/tmp/data")
	require.Equal(t, filepath.Join("/tmp/data", "grammar"), resolveGrammarDir(opts))
}

func TestParseOptions_DataDirAndDoctor(t *testing.T) {
	p := &CommandLineParser{}
	opts, err := p.ParseOptions([]string{"--data-dir", "/data", "--doctor"})
	require.NoError(t, err)
	require.Equal(t, "/data", opts.DataDir)
	require.True(t, opts.PrintDoctor)
}

func TestLintDefaultAutoDetect(t *testing.T) {
	// product lint without -l enables auto-detect
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"lint", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "This is an test.", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.True(t, opts.IsAutoDetect() || opts.Language != "", "expected auto or lang")
			// still run core
			return CoreCheckHook(w, text, opts)
		},
	}, &out, &errb)
	// may be 0 or 1 depending on detection + matches
	require.True(t, code == 0 || code == 1, "code=%d err=%s", code, errb.String())
	_ = runtime.Version()
}
