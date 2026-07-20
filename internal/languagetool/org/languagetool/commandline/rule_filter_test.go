package commandline

import (
	"bytes"
	"io"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestFilterMatchesByRules(t *testing.T) {
	ms := []*rules.RuleMatch{
		rules.NewRuleMatch(idR{"A"}, nil, 0, 1, "a"),
		rules.NewRuleMatch(idR{"B"}, nil, 0, 1, "b"),
		rules.NewRuleMatch(idR{"C"}, nil, 0, 1, "c"),
	}
	got := FilterMatchesByRules(ms, []string{"B"}, nil, false)
	require.Len(t, got, 2)
	require.Equal(t, "A", ruleIDOfMatch(got[0]))
	require.Equal(t, "C", ruleIDOfMatch(got[1]))

	// -e without -eo: Java enableRule only — does not hide other matches
	got2 := FilterMatchesByRules(ms, nil, []string{"A", "C"}, false)
	require.Len(t, got2, 3)
	// -e with -eo: only listed IDs
	got3 := FilterMatchesByRules(ms, nil, []string{"A", "C"}, true)
	require.Len(t, got3, 2)
}

func TestMain_EnglishFileRuleDisabled(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-d", "A,B", "f.txt"}, RunHooks{
		ReadFile: func(path string) (string, error) { return "text", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.ElementsMatch(t, []string{"A", "B"}, opts.GetDisabledRules())
			ms := []*rules.RuleMatch{
				rules.NewRuleMatch(idR{"A"}, nil, 0, 1, "a"),
				rules.NewRuleMatch(idR{"C"}, nil, 0, 1, "c"),
			}
			filtered := FilterMatchesByRules(ms, opts.GetDisabledRules(), opts.GetEnabledRules(), opts.IsUseEnabledOnly())
			for _, m := range filtered {
				_, _ = io.WriteString(w, ruleIDOfMatch(m)+"\n")
			}
			return len(filtered), nil
		},
	}, &out, &errb)
	require.Equal(t, 2, code)
	require.Contains(t, out.String(), "C")
	require.NotContains(t, out.String(), "A\n")
}

func TestMain_EnglishFileRuleEnabled(t *testing.T) {
	var out, errb bytes.Buffer
	// -e alone does not use enabledOnly (Java -eo required to hide others)
	code := RunWithIO([]string{"-l", "en", "-e", "ONLY", "f.txt"}, RunHooks{
		ReadFile: func(path string) (string, error) { return "text", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.Contains(t, opts.GetEnabledRules(), "ONLY")
			require.False(t, opts.IsUseEnabledOnly())
			ms := []*rules.RuleMatch{
				rules.NewRuleMatch(idR{"ONLY"}, nil, 0, 1, "o"),
				rules.NewRuleMatch(idR{"OTHER"}, nil, 0, 1, "x"),
			}
			filtered := FilterMatchesByRules(ms, opts.GetDisabledRules(), opts.GetEnabledRules(), opts.IsUseEnabledOnly())
			return len(filtered), nil
		},
	}, &out, &errb)
	require.Equal(t, 2, code) // matches found (exit 2); both still reported without -eo
}

func TestMain_EnglishFileRuleEnabledOnly(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-e", "ONLY", "-eo", "f.txt"}, RunHooks{
		ReadFile: func(path string) (string, error) { return "text", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.True(t, opts.IsUseEnabledOnly())
			ms := []*rules.RuleMatch{
				rules.NewRuleMatch(idR{"ONLY"}, nil, 0, 1, "o"),
				rules.NewRuleMatch(idR{"OTHER"}, nil, 0, 1, "x"),
			}
			filtered := FilterMatchesByRules(ms, opts.GetDisabledRules(), opts.GetEnabledRules(), opts.IsUseEnabledOnly())
			require.Len(t, filtered, 1)
			require.Equal(t, "ONLY", ruleIDOfMatch(filtered[0]))
			return len(filtered), nil
		},
	}, &out, &errb)
	require.Equal(t, 2, code) // one match remains → exit 2
}

func TestMain_EnglishFileFakeRuleEnabled(t *testing.T) {
	var out, errb bytes.Buffer
	// -e alone: enabling a non-existent id does not hide other active matches
	code := RunWithIO([]string{"-l", "en", "-e", "DOES_NOT_EXIST", "f.txt"}, RunHooks{
		ReadFile: func(path string) (string, error) { return "text", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			ms := []*rules.RuleMatch{rules.NewRuleMatch(idR{"REAL"}, nil, 0, 1, "r")}
			filtered := FilterMatchesByRules(ms, nil, opts.GetEnabledRules(), opts.IsUseEnabledOnly())
			require.Len(t, filtered, 1)
			return len(filtered), nil
		},
	}, &out, &errb)
	require.Equal(t, 2, code) // REAL still reported
}
