package commandline

// Twin of languagetool-commandline/src/test/java/org/languagetool/commandline/MainTest.java
//
// Full end-to-end CLI with language modules deferred; green smokes cover RunWithIO surface
// (usage, languages, stdin check/json/tagger, exit codes). Integration cases soft-skipped.
import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Port of MainTest.testUsageMessage
func TestMain_UsageMessage(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"--help"}, RunHooks{}, &out, &errb)
	require.Equal(t, 0, code)
	require.Contains(t, out.String(), "Usage:")
	require.Contains(t, out.String(), "--language")
	require.Contains(t, out.String(), "--json")
}

// Port of MainTest.testPrintLanguages
func TestMain_PrintLanguages(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"--list"}, RunHooks{
		ListLanguages: func(w io.Writer) error {
			_, _ = io.WriteString(w, "en-US\nde-DE\nuk-UA\n")
			return nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
	require.Contains(t, out.String(), "en-US")
	require.Contains(t, out.String(), "uk-UA")
}

// Port of MainTest.testEnglishStdIn1 — stdin check via core rule pack
func TestMain_EnglishStdIn1(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "This is an test.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.Equal(t, 2, code) // matches found
	require.Contains(t, out.String(), "EN_A_VS_AN")
}

// Port of MainTest.testEnglishStdInJsonOutput
func TestMain_EnglishStdInJsonOutput(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "ok", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.Equal(t, OutputJSON, opts.OutputFormat)
			_, _ = io.WriteString(w, `{"matches":[]}`)
			return 0, nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
	require.Contains(t, out.String(), `"matches"`)
}

// Port of MainTest.testEnglishTagger
func TestMain_EnglishTagger(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-t", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Hello world", nil },
		Tag: func(w io.Writer, text string, opts *CommandLineOptions) error {
			require.True(t, opts.TaggerOnly)
			_, _ = io.WriteString(w, FormatTagLine(text, strings.Fields(text))+"\n")
			return nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
	require.Contains(t, out.String(), "Hello")
}

// Port of MainTest.testEnglishFile — ReadFile hook
func TestMain_EnglishFile(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "sample.txt"}, RunHooks{
		ReadFile: func(path string) (string, error) {
			require.Equal(t, "sample.txt", path)
			return "clean text", nil
		},
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.Equal(t, "clean text", text)
			return 0, nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
}

// Port of MainTest.testLangWithCountryVariant
func TestMain_LangWithCountryVariant(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en-US", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "x", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.Equal(t, "en-US", opts.Language)
			return 0, nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
}

// Remaining Java CLI integration cases need full language modules / bitext / XML filter pipeline.
func TestMain_EnglishFileApplySuggestions(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "sample.txt"}, RunHooks{
		ReadFile: func(path string) (string, error) {
			return "I've have a book", nil
		},
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.True(t, opts.IsApplySuggestions())
			// synthetic match for "have" → "had"
			m := rules.NewRuleMatch(nil, nil, 5, 9, "tense")
			m.SetSuggestedReplacement("had")
			return ApplySuggestionsCheck(w, text, []*rules.RuleMatch{m})
		},
	}, &out, &errb)
	require.Equal(t, 2, code) // one match found
	require.Contains(t, out.String(), "I've had")
}
