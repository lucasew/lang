package commandline

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestMatchesAsSARIF(t *testing.T) {
	m := rules.NewRuleMatch(nil, nil, 8, 10, `Use "a" before a consonant sound`)
	m.Rule = &rules.BaseRule{ID: "EN_A_VS_AN"}
	s := MatchesAsSARIF([]*rules.RuleMatch{m}, "This is an test.", "sample.txt", "en")
	require.Contains(t, s, `"version":"2.1.0"`)
	require.Contains(t, s, "EN_A_VS_AN")
	require.Contains(t, s, `"level":"error"`)
	require.Contains(t, s, `"type":"grammar"`)
	require.Contains(t, s, `"helpUri":"https://community.languagetool.org/rule/show/EN_A_VS_AN?lang=en"`)
	var raw map[string]any
	require.NoError(t, json.Unmarshal([]byte(s), &raw))
}

func TestCoreCheckHook_SARIF(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--sarif", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "This is an test.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.Equal(t, 1, code) // SPEC: error-severity findings
	body := out.String()
	require.Contains(t, body, `"version":"2.1.0"`)
	require.Contains(t, body, "EN_A_VS_AN")
	require.Contains(t, body, `"level":"error"`)
	_ = io.Discard
}

func TestCoreCheckHook_DisableCategories(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--disablecategories", "GRAMMAR", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "This is an test.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.Equal(t, 0, code)
	require.NotContains(t, out.String(), "EN_A_VS_AN")
}

func TestParseOptions_SARIF(t *testing.T) {
	p := &CommandLineParser{}
	opts, err := p.ParseOptions([]string{"--sarif", "-l", "en", "-"})
	require.NoError(t, err)
	require.Equal(t, OutputSARIF, opts.OutputFormat)
}
