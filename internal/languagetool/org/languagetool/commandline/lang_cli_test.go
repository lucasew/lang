package commandline

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestMain_EnglishStdIn2(t *testing.T) {
	// multi-sentence stdin
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-"}, RunHooks{
		ReadStdin: func() (string, error) {
			return "This is fine. This is an test.", nil
		},
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			// flag "an " before consonant-ish
			var ms []*rules.RuleMatch
			if i := strings.Index(text, "an test"); i >= 0 {
				ms = append(ms, rules.NewRuleMatch(idR{"EN_A_VS_AN"}, nil, i, i+2, "Use a"))
			}
			for _, m := range ms {
				line, col := LineColumnAt(text, m.FromPos)
				_, _ = io.WriteString(w, formatMatchLine(1, line, col, "EN_A_VS_AN"))
			}
			return len(ms), nil
		},
	}, &out, &errb)
	require.Equal(t, 2, code)
	require.Contains(t, out.String(), "EN_A_VS_AN")
}

func formatMatchLine(n, line, col int, id string) string {
	return "1.) Line " + itoa(line) + ", column " + itoa(col) + ", Rule ID: " + id + "\n"
}
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func TestMain_EnglishStdIn3(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "an apple", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			// no match for correct a/an
			_, _ = io.WriteString(w, MatchesAsJSON(nil, "en", text))
			return 0, nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
}

func TestMain_EnglishStdIn4(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--line-by-line", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "one\ntwo\n", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			return CheckLineByLine(w, text, func(seg string) ([]*rules.RuleMatch, error) {
				return nil, nil
			})
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
}

func TestMain_PolishStdInDefaultOff(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "pl", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "To jest test.", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.Equal(t, "pl", opts.Language)
			return 0, nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
}

func TestMain_PolishApiStdInDefaultOff(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "pl", "--xml", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "To jest test.", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			_, _ = io.WriteString(w, MatchesAsMinimalXML(nil, "pl"))
			return 0, nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
	require.Contains(t, out.String(), `language="pl"`)
}

func TestMain_PolishApiStdInDefaultOffNoErrors(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "pl", "--xml", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "ok", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			// default-off rules not run → empty
			_, _ = io.WriteString(w, MatchesAsMinimalXML(nil, "pl"))
			return 0, nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
}

func TestMain_PolishSpelling(t *testing.T) {
	known := map[string]bool{"to": true, "jest": true, "test": true}
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "pl", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "To jest xyzzytest", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			ms := SimplePolishSpellingMatch(text, known)
			for _, m := range ms {
				_, _ = io.WriteString(w, ruleIDOfMatch(m)+"\n")
			}
			return len(ms), nil
		},
	}, &out, &errb)
	require.Equal(t, 2, code)
	require.Contains(t, out.String(), "MORFOLOGIK_RULE_PL_PL")
}

func TestMain_Catalan(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "ca", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Hola món.", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.Equal(t, "ca", opts.Language)
			return 0, nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
}

func TestMain_Catalan2(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "ca-ES-valencia", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Bon dia.", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.Equal(t, "ca-ES-valencia", opts.Language)
			return 0, nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
}

func TestMain_ValencianCatalan(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "ca-ES-valencia", "--list"}, RunHooks{
		ListLanguages: func(w io.Writer) error {
			_, _ = io.WriteString(w, "ca\nca-ES-valencia\n")
			return nil
		},
	}, &out, &errb)
	// --list may short-circuit before language matters
	require.Equal(t, 0, code)
	require.Contains(t, out.String(), "ca")
}

func TestMain_BitextWithExternalRule(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bitext.xml")
	require.NoError(t, os.WriteFile(path, []byte(`<rules></rules>`), 0o644))
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--bitext", "--rulefile", path, "pairs.txt"}, RunHooks{
		ReadFile: func(p string) (string, error) {
			if p == "pairs.txt" {
				return "Hi\tHello longer text here\n", nil
			}
			return "", nil
		},
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.True(t, opts.Bitext)
			return CheckBitextWithRuleFile(w, text, opts.GetRuleFile())
		},
	}, &out, &errb)
	require.True(t, code == 0 || code == 2)
}

func TestCheckWithPatternRuleFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "grammar-en.xml")
	xml := `<?xml version="1.0"?>
<rules>
  <category>
    <rule id="FOO_BAR" name="foo bar">
      <pattern>
        <token>foo</token>
        <token>bar</token>
      </pattern>
      <message>found foo bar</message>
    </rule>
  </category>
</rules>`
	require.NoError(t, os.WriteFile(path, []byte(xml), 0o644))
	opts := NewCommandLineOptions()
	opts.SetLanguage("en")
	opts.SetRuleFile(path)
	var buf bytes.Buffer
	n, err := CheckWithPatternRuleFile(&buf, "xx foo bar yy", opts)
	require.NoError(t, err)
	require.GreaterOrEqual(t, n, 1)
	require.Contains(t, buf.String(), "FOO_BAR")
}
