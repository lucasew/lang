package commandline

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestWriteLintMatches(t *testing.T) {
	m := rules.NewRuleMatch(nil, nil, 8, 10, `Use "a" before a consonant sound`)
	m.Rule = &rules.BaseRule{ID: "EN_A_VS_AN"}
	m.SetSuggestedReplacements([]string{"a"})
	var buf bytes.Buffer
	require.NoError(t, WriteLintMatches(&buf, []*rules.RuleMatch{m}, "This is an test.", "doc.txt"))
	out := buf.String()
	require.Contains(t, out, "location")
	require.Contains(t, out, "severity")
	require.Contains(t, out, "EN_A_VS_AN")
	require.Contains(t, out, "error")
	require.Contains(t, out, "grammar")
	require.Contains(t, out, "doc.txt:")
	require.Contains(t, out, "a")
}

func TestCoreCheckHook_Lint(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--lint", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "This is an test.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.Equal(t, 1, code) // SPEC: error-severity findings
	body := out.String()
	require.True(t, strings.Contains(body, "EN_A_VS_AN"), body)
	require.True(t, strings.Contains(body, "error"), body)
	require.True(t, strings.Contains(body, "grammar"), body)
	require.True(t, strings.Contains(body, "stdin:"), body)
	_ = io.Discard
}

func TestParseOptions_Lint(t *testing.T) {
	p := &CommandLineParser{}
	opts, err := p.ParseOptions([]string{"--lint", "-l", "en", "-"})
	require.NoError(t, err)
	require.Equal(t, OutputLint, opts.OutputFormat)
}
