package commandline

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestMatchesAsMinimalXML(t *testing.T) {
	m := rules.NewRuleMatch(idR{"EN_A_VS_AN"}, nil, 8, 10, "Use a instead of an")
	m.SetSuggestedReplacement("a")
	xml := MatchesAsMinimalXML([]*rules.RuleMatch{m}, "en")
	require.Contains(t, xml, `ruleId="EN_A_VS_AN"`)
	require.Contains(t, xml, `fromx="8"`)
	require.Contains(t, xml, "replacements=")
	require.Contains(t, xml, `language="en"`)
}

func TestMatchesAsJSON(t *testing.T) {
	m := rules.NewRuleMatch(idR{"R1"}, nil, 0, 3, "msg")
	js := MatchesAsJSON([]*rules.RuleMatch{m}, "en", "abc")
	require.Contains(t, js, "R1")
	require.True(t, len(js) > 2)
}

func TestAPILineColumnAt(t *testing.T) {
	text := "ab\ncde\nf"
	// offset 0 → 1,1
	l, c := LineColumnAt(text, 0)
	require.Equal(t, 1, l)
	require.Equal(t, 1, c)
	// after newline at index 3 ('c')
	l, c = LineColumnAt(text, 3)
	require.Equal(t, 2, l)
	require.Equal(t, 1, c)
}

func TestMain_EnglishFileAPI(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--xml", "f.txt"}, RunHooks{
		ReadFile: func(path string) (string, error) { return "This is an test.", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.Equal(t, OutputXML, opts.OutputFormat)
			m := rules.NewRuleMatch(idR{"EN_A_VS_AN"}, nil, 8, 10, "Use 'a'")
			xml := MatchesAsMinimalXML([]*rules.RuleMatch{m}, opts.Language)
			_, _ = io.WriteString(w, xml)
			return 1, nil
		},
	}, &out, &errb)
	require.Equal(t, 2, code)
	require.Contains(t, out.String(), "<matches")
	require.Contains(t, out.String(), "EN_A_VS_AN")
}

func TestMain_PolishFileAPI(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "pl", "--xml", "f.txt"}, RunHooks{
		ReadFile: func(path string) (string, error) { return "To jest test.", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.Equal(t, "pl", opts.Language)
			_, _ = io.WriteString(w, MatchesAsMinimalXML(nil, "pl"))
			return 0, nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
	require.Contains(t, out.String(), `language="pl"`)
}

func TestMain_PolishLineNumbers(t *testing.T) {
	text := "line1\nline2 bad\n"
	// 'b' of bad is at offset after "line1\nline2 " = 6+6=12? line1\n = 6, line2 = 5, space=1 → 12
	offset := 12
	l, c := LineColumnAt(text, offset)
	require.Equal(t, 2, l)
	require.GreaterOrEqual(t, c, 1)
}

func TestMain_EnglishFileAutoDetect(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"--autoDetect", "f.txt"}, RunHooks{
		ReadFile: func(path string) (string, error) { return "Das ist ein Test mit Größe.", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.True(t, opts.IsAutoDetect())
			lang := ResolveLanguage(text, opts, nil)
			require.Equal(t, "de", lang)
			_, _ = io.WriteString(w, "lang="+lang+"\n")
			return 0, nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
	require.Contains(t, out.String(), "lang=de")
}

func TestMain_EnglishStdInAutoDetect(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-adl", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Hello world from English text.", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			lang := ResolveLanguage(text, opts, nil)
			require.Equal(t, "en", lang)
			return 0, nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
}

func TestMain_FileWithExternalRule(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "grammar-en.xml")
	require.NoError(t, os.WriteFile(path, []byte(`<rules lang="en"><rule id="X"><pattern><token>foo</token></pattern></rule></rules>`), 0o644))
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--rulefile", path, "f.txt"}, RunHooks{
		ReadFile: func(p string) (string, error) {
			if p == path {
				// not used for input
			}
			return "foo bar", nil
		},
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.Equal(t, path, opts.GetRuleFile())
			content, err := LoadRuleFile(opts.GetRuleFile())
			require.NoError(t, err)
			require.Contains(t, content, "rule id=")
			require.Equal(t, "en", InferLanguageFromRuleFileName(path))
			return 0, nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
}

func TestMain_GermanFileWithURL(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "de", "f.txt"}, RunHooks{
		ReadFile: func(path string) (string, error) { return "text", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			m := rules.NewRuleMatch(idR{"DE_URL"}, nil, 0, 4, "msg")
			m.SetURL("https://example.com/rule")
			require.Equal(t, "https://example.com/rule", m.GetURL())
			_, _ = io.WriteString(w, "URL: "+m.GetURL()+"\n")
			return 1, nil
		},
	}, &out, &errb)
	require.Equal(t, 2, code)
	require.Contains(t, out.String(), "https://example.com/rule")
}

func TestMain_EnglishFileVerbose(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-v", "f.txt"}, RunHooks{
		ReadFile: func(path string) (string, error) { return "ok", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.True(t, opts.Verbose)
			return 0, nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
}
