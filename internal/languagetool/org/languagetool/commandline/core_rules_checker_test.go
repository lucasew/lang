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
