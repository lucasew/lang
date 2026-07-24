package commandline

import (
	"bytes"
	"io"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

type idR struct{ id string }

func (r idR) GetID() string { return r.id }

func TestSplitParagraphs(t *testing.T) {
	// Java Main.runOnFile: each readLine is appended with '\n'; empty line breaks.
	// Input lines: "P1 line a", "P1 line b", "", "P2 only"
	// → para0 "P1 line a\nP1 line b\n\n", para1 "P2 only\n"
	text := "P1 line a\nP1 line b\n\nP2 only\n"
	paras := SplitParagraphs(text, false)
	require.Len(t, paras, 2)
	require.Equal(t, "P1 line a\nP1 line b\n\n", paras[0])
	require.Equal(t, "P2 only\n", paras[1])

	// singleLineBreakMarksPara: every line is a breakpoint after append
	paras2 := SplitParagraphs("A\nB\n\nC\n", true)
	require.Equal(t, []string{"A\n", "B\n", "\n", "C\n"}, paras2)
}

func TestCheckLineByLine(t *testing.T) {
	var buf bytes.Buffer
	n, err := CheckLineByLine(&buf, "good line\nbad line\n", func(seg string) ([]*rules.RuleMatch, error) {
		if seg == "bad line" {
			m := rules.NewRuleMatch(idR{"DEMO"}, nil, 0, 3, "bad")
			return []*rules.RuleMatch{m}, nil
		}
		return nil, nil
	})
	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.Contains(t, buf.String(), "Line 2")
	require.Contains(t, buf.String(), "DEMO")
}

func TestCheckParagraphs(t *testing.T) {
	var buf bytes.Buffer
	text := "First para has error.\n\nSecond clean."
	n, err := CheckParagraphs(&buf, text, false, func(seg string) ([]*rules.RuleMatch, error) {
		if containsWord(seg, "error") {
			return []*rules.RuleMatch{rules.NewRuleMatch(idR{"ERR"}, nil, 0, 5, "e")}, nil
		}
		return nil, nil
	})
	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.Contains(t, buf.String(), "Paragraph 1")
}

func containsWord(s, w string) bool {
	return len(s) >= len(w) && (s == w || len(s) > 0 && (stringContains(s, w)))
}
func stringContains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (s == sub || len(s) > 0 && indexOf(s, sub) >= 0))
}
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestMain_EnglishLineMode(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--line-by-line", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "ok\nbad\n", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.True(t, opts.LineByLine)
			return CheckLineByLine(w, text, func(seg string) ([]*rules.RuleMatch, error) {
				if seg == "bad" {
					return []*rules.RuleMatch{rules.NewRuleMatch(idR{"X"}, nil, 0, 3, "m")}, nil
				}
				return nil, nil
			})
		},
	}, &out, &errb)
	require.Equal(t, 2, code)
	require.Contains(t, out.String(), "Line 2")
}

func TestMain_EnglishParaMode(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Para one.\n\nPara two.", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			// paragraph mode via SplitParagraphs
			return CheckParagraphs(w, text, false, func(seg string) ([]*rules.RuleMatch, error) {
				return nil, nil
			})
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
}
