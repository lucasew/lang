package commandline

import (
	"bytes"
	"io"
	"os"
	"runtime"
	"path/filepath"
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
	body := out.String()
	require.Contains(t, body, "EN_A_VS_AN")
	require.Contains(t, body, "matches")
	require.Contains(t, body, `"issueType":"grammar"`)
	require.Contains(t, body, `"severity":"error"`)
	require.Contains(t, body, `"id":"GRAMMAR"`)
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

func TestCoreCheckHook_FalseFriends(t *testing.T) {
	dir := t.TempDir()
	ff := dir + "/ff.xml"
	require.NoError(t, os.WriteFile(ff, []byte(`<?xml version="1.0"?>
<rules>
  <rulegroup id="ABILITY">
    <rule>
      <pattern lang="en">
        <token>ability</token>
      </pattern>
      <translation lang="fr">aptitude</translation>
    </rule>
  </rulegroup>
</rules>`), 0o644))

	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-m", "fr", "--falsefriends", ff, "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "My ability is great.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.Equal(t, 2, code, errb.String())
	require.Contains(t, out.String(), "ABILITY")
	require.Contains(t, out.String(), "aptitude")
}

func TestCoreCheckHook_LineByLine(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--line-by-line", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "All good.\nThis is an test.\n", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.Equal(t, 2, code)
	require.Contains(t, out.String(), "EN_A_VS_AN")
}

func TestCoreRulesChecker_CleanOverlaps(t *testing.T) {
	c := NewCoreRulesChecker("en")
	c.CleanOverlaps = true
	// synthetic overlapping: a/an and others on same region rare; just ensure no panic
	ms, err := c.Check("This is an test.")
	require.NoError(t, err)
	require.NotEmpty(t, ms)
}

func TestCoreCheckHook_PickyLevel(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--level", "picky", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "I have alot of work.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.Equal(t, 2, code, errb.String())
	require.Contains(t, out.String(), "EN_A_LOT")
}

func TestCoreCheckHook_GrammarDir(t *testing.T) {
	// module-relative testdata path
	dir := filepath.Join("..", "..", "..", "..", "..", "testdata", "grammar")
	// resolve from commandline package: internal/languagetool/org/languagetool/commandline = 5 ups
	t.Setenv("LANG_GRAMMAR_DIR", filepath.Clean(filepath.Join("testdata", "grammar")))
	// From package dir during test, cwd is package dir — use absolute from runtime
	_, file, _, _ := runtime.Caller(0)
	// commandline → languagetool → org → languagetool → internal → module root (5)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../.."))
	t.Setenv("LANG_GRAMMAR_DIR", filepath.Join(root, "testdata", "grammar"))

	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Well, your welcome to try.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.Equal(t, 2, code, errb.String()+" cwd grammar: "+os.Getenv("LANG_GRAMMAR_DIR"))
	require.Contains(t, out.String(), "EN_SOFT_YOUR_YOU_RE")
	_ = dir
}
