package commandline

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCoreRulesChecker_Check(t *testing.T) {
	c := NewCoreRulesChecker("en")
	ms, err := c.Check("This is an test.")
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	require.Equal(t, "EN_A_VS_AN", ms[0].Rule.(interface{ GetID() string }).GetID())
}

func TestCoreCheckHook_StdIn(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "This is an test.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.Equal(t, 2, code) // matches found
	require.Contains(t, out.String(), "EN_A_VS_AN")
}

func TestCoreCheckHook_JSON(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "This is an test.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.Equal(t, 2, code)
	require.Contains(t, out.String(), "EN_A_VS_AN")
	require.Contains(t, out.String(), "matches")
}

func TestCoreApplySuggestionsHook(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "This is an test.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	// apply path uses Check hook slot; exit code depends on match count
	require.True(t, code == 0 || code == 2)
	require.Contains(t, strings.TrimSpace(out.String()), "This is a test.")
}

func TestCoreRulesChecker_German(t *testing.T) {
	c := NewCoreRulesChecker("de-DE")
	ms, err := c.Check("Ein Test Test.")
	require.NoError(t, err)
	require.NotEmpty(t, ms)
}

func TestCoreCheckHook_NoMatches(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "All good here.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.Equal(t, 0, code)
	_ = io.Discard
}

func TestDefaultCoreHooks_ListAndTag(t *testing.T) {
	hooks := DefaultCoreHooks()
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"--list"}, hooks, &out, &errb)
	require.Equal(t, 0, code)
	require.Contains(t, out.String(), "en-US")
	require.Contains(t, out.String(), "de-DE")

	out.Reset()
	errb.Reset()
	code = RunWithIO([]string{"-l", "en", "-t", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Hello world", nil },
		Tag:       hooks.Tag,
	}, &out, &errb)
	require.Equal(t, 0, code)
	require.Contains(t, out.String(), "Hello")
}

func TestCoreCheckHook_DisableRule(t *testing.T) {
	var out, errb bytes.Buffer
	// disable a/an — should not report EN_A_VS_AN
	code := RunWithIO([]string{"-l", "en", "-d", "EN_A_VS_AN", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "This is an test.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	// may still match uppercase etc.; must not include EN_A_VS_AN
	require.NotContains(t, out.String(), "EN_A_VS_AN")
	_ = code
}

func TestCoreCheckHook_EnabledOnly(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-e", "EN_A_VS_AN", "--enabledonly", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "This is an test. hello  world", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.Equal(t, 2, code)
	require.Contains(t, out.String(), "EN_A_VS_AN")
	require.NotContains(t, out.String(), "WHITESPACE_RULE")
}

func TestApplyCLIRuleFilters(t *testing.T) {
	c := NewCoreRulesChecker("en")
	opts := NewCommandLineOptions()
	opts.SetDisabledRules([]string{"EN_A_VS_AN"})
	ApplyCLIRuleFilters(c.LT(), opts)
	ms, err := c.Check("This is an test.")
	require.NoError(t, err)
	for _, m := range ms {
		if g, ok := m.Rule.(interface{ GetID() string }); ok {
			require.NotEqual(t, "EN_A_VS_AN", g.GetID())
		}
	}
}
