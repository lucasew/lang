package commandline

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCoreCheckHook_LintExitUsesErrorSeverity(t *testing.T) {
	// style-only long sentence with low threshold → severity note → exit 0 from RunWithIO
	var out, errb bytes.Buffer
	words := strings.Repeat("word ", 12)
	text := strings.TrimSpace(words) + "."
	code := RunWithIO([]string{"-l", "en", "--lint", "--ruleValues", "TOO_LONG_SENTENCE:3", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return text, nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	// only style matches → error count 0 → exit 0
	require.Equal(t, 0, code, out.String()+errb.String())
	require.Contains(t, out.String(), "TOO_LONG_SENTENCE")
	_ = io.Discard
}

func TestCoreCheckHook_LintGrammarExit1(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--lint", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "This is an test.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.Equal(t, 1, code) // SPEC: error severity
	require.Contains(t, out.String(), "EN_A_VS_AN")
}

func TestParseOptions_RuleValues(t *testing.T) {
	p := &CommandLineParser{}
	opts, err := p.ParseOptions([]string{"--ruleValues", "TOO_LONG_SENTENCE:8", "-l", "en", "-"})
	require.NoError(t, err)
	require.Equal(t, []string{"TOO_LONG_SENTENCE:8"}, opts.RuleValues)
}

func TestParseRuleValues(t *testing.T) {
	m := parseRuleValues([]string{"A:1,B:2"})
	require.Equal(t, "1", m["A"])
	require.Equal(t, "2", m["B"])
}
