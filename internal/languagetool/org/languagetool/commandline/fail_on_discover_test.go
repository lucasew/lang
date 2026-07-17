package commandline

import (
	"bytes"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDetectLanguageHeuristic_Expanded(t *testing.T) {
	require.Equal(t, "de", DetectLanguageHeuristic("Das ist ein schönes Buch."))
	require.Equal(t, "es", DetectLanguageHeuristic("¿Dónde está la casa?"))
	require.Equal(t, "pl", DetectLanguageHeuristic("Zażółć gęślą jaźń"))
	require.Equal(t, "uk", DetectLanguageHeuristic("Привіт, це українська мова з ї."))
	require.Equal(t, "ru", DetectLanguageHeuristic("Привет мир"))
	require.Equal(t, "el", DetectLanguageHeuristic("Καλημέρα κόσμε"))
	require.Equal(t, "en", DetectLanguageHeuristic("Hello world"))
}

func TestWalkUpFind_Testdata(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	// start from this package dir
	start := filepath.Dir(file)
	found := WalkUpFind(start, filepath.Join("testdata", "grammar"))
	require.NotEmpty(t, found)
	require.Contains(t, found, "testdata")
	require.Contains(t, found, "grammar")
}

func TestFailOnWarning_LongSentence(t *testing.T) {
	// style long sentence is note by SoftRuleMeta; fail-on=warning should still exit 0 for note-only
	// Capitalize start to avoid UPPERCASE_SENTENCE_START (warning).
	var out, errb bytes.Buffer
	words := "Word " + strings.Repeat("word ", 11)
	text := strings.TrimSpace(words) + "."
	code := RunWithIO([]string{"-l", "en", "--lint", "--fail-on", "warning", "--ruleValues", "TOO_LONG_SENTENCE:3", "-d", "UPPERCASE_SENTENCE_START", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return text, nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.Equal(t, 0, code, out.String()+errb.String())
	require.Contains(t, out.String(), "TOO_LONG_SENTENCE")
}

func TestFailOnNote_LongSentence(t *testing.T) {
	var out, errb bytes.Buffer
	words := "Word " + strings.Repeat("word ", 11)
	text := strings.TrimSpace(words) + "."
	code := RunWithIO([]string{"-l", "en", "--lint", "--fail-on", "note", "--ruleValues", "TOO_LONG_SENTENCE:3", "-d", "UPPERCASE_SENTENCE_START", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return text, nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.Equal(t, 1, code, out.String()+errb.String())
}

func TestParseOptions_FailOn(t *testing.T) {
	p := &CommandLineParser{}
	opts, err := p.ParseOptions([]string{"--fail-on", "warning", "--lint", "-"})
	require.NoError(t, err)
	require.Equal(t, "warning", opts.FailOn)
	require.Equal(t, "warning", opts.GetFailOn())
}
