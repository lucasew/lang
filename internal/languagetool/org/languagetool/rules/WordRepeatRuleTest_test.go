package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.WordRepeatRuleTest — full-strength asserts.

func TestWordRepeatRule_Test(t *testing.T) {
	rule := NewWordRepeatRule(map[string]string{
		"repetition":            "Word repetition",
		"desc_repetition":       "Word repetition",
		"desc_repetition_short": "repetition",
	})

	assertGood := func(s string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(s))
		require.Equal(t, 0, len(matches), "assertGood %q", s)
	}
	assertBad := func(s string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(s))
		require.Equal(t, 1, len(matches), "assertBad %q", s)
	}

	assertGood("A test")
	assertGood("A test.")
	assertGood("A test...")
	assertGood("1 000 000 years")
	assertGood("010 020 030")
	// thumbs up, green heart, evergreen tree x2 as emoji
	assertGood("👍💚🌲🌲")

	assertBad("A A test")
	assertBad("A a test")
	assertBad("This is is a test")
}
