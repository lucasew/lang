package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

// applyOrdinal runs AcceptRuleMatch with a single suggested replacement and
// returns the rewritten first suggestion (Java-visible outcome).
func applyOrdinal(t *testing.T, suggestion string) string {
	t.Helper()
	f := NewOrdinalSuffixFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("O"), nil, 0, len(suggestion), "msg")
	m.SetSuggestedReplacements([]string{suggestion})
	out := f.AcceptRuleMatch(m, nil, 0, nil, nil)
	require.Same(t, m, out, "Java returns the same RuleMatch instance")
	reps := out.GetSuggestedReplacements()
	require.Len(t, reps, 1)
	return reps[0]
}

func TestOrdinalSuffixFilter_AcceptRuleMatch_Suffixes(t *testing.T) {
	// Java control flow: NON_DIGITS strip → teens PATTERN.matches → else endsWith 1/2/3/else.
	cases := []struct {
		in, want string
	}{
		// broken suggestions → correct ordinal
		{"1nd", "1st"},
		{"2th", "2nd"},
		{"3st", "3rd"},
		{"4nd", "4th"},
		{"5st", "5th"},
		{"11st", "11th"},
		{"12nd", "12th"},
		{"13rd", "13th"},
		{"21nd", "21st"},
		{"22th", "22nd"},
		{"23th", "23rd"},
		// teens: full-string PATTERN.matches (ends with 11|12|13) — not substring search
		{"111st", "111th"},
		{"112nd", "112th"},
		{"113rd", "113th"},
		{"211nd", "211th"},
		// must NOT treat internal 11/12/13 as teens (Java Matcher.matches is full match)
		{"121nd", "121st"},
		{"123st", "123rd"},
		{"2115th", "2115th"},
		// NON_DIGITS strip of first suggestion
		{"1st!", "1st"},
		{"#22nd", "22nd"},
		{"the 3rd", "3rd"},
		{"", "th"}, // no digits → empty ordinal + "th"
	}
	for _, tc := range cases {
		t.Run(tc.in+"→"+tc.want, func(t *testing.T) {
			require.Equal(t, tc.want, applyOrdinal(t, tc.in))
		})
	}
}

func TestOrdinalSuffixFilter_AcceptRuleMatch_MutatesMatch(t *testing.T) {
	f := NewOrdinalSuffixFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("O"), nil, 0, 3, "msg")
	m.SetSuggestedReplacements([]string{"1nd"})
	out := f.AcceptRuleMatch(m, nil, 0, nil, nil)
	require.Same(t, m, out)
	require.Equal(t, []string{"1st"}, out.GetSuggestedReplacements())
	// Java setSuggestedReplacement replaces the whole list with one entry.
	require.Equal(t, []string{"1st"}, m.GetSuggestedReplacements())

	m2 := rules.NewRuleMatch(rules.NewFakeRule("O"), nil, 0, 3, "msg")
	m2.SetSuggestedReplacements([]string{"22th", "ignored-second"})
	out = f.AcceptRuleMatch(m2, nil, 0, nil, nil)
	// Only first suggestion is read; setSuggestedReplacement collapses to one.
	require.Equal(t, []string{"22nd"}, out.GetSuggestedReplacements())
}

func TestOrdinalSuffixFilter_EmptySuggestionsPanics(t *testing.T) {
	// Java: getSuggestedReplacements().get(0) → IndexOutOfBoundsException
	f := NewOrdinalSuffixFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("O"), nil, 0, 1, "msg")
	// no replacements set
	require.Panics(t, func() {
		f.AcceptRuleMatch(m, nil, 0, nil, nil)
	})
}

func TestOrdinalSuffixFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.en.OrdinalSuffixFilter"))
	f := patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.en.OrdinalSuffixFilter")
	require.NotNil(t, f)
	// Factory yields a working filter (AcceptRuleMatch rewrite).
	m := rules.NewRuleMatch(rules.NewFakeRule("O"), nil, 0, 2, "msg")
	m.SetSuggestedReplacements([]string{"1nd"})
	out := f.AcceptRuleMatch(m, nil, 0, nil, nil)
	require.Equal(t, []string{"1st"}, out.GetSuggestedReplacements())
}
