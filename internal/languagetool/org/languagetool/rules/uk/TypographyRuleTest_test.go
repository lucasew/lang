package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TypographyRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestTypographyRule_Rule(t *testing.T) {
	rule := NewTypographyRule(nil)
	assert0 := func(s string) {
		t.Helper()
		require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(s))), "good %q", s)
	}
	assert1 := func(s string, suggs ...string) {
		t.Helper()
		m := rule.Match(languagetool.AnalyzePlain(s))
		require.Equal(t, 1, len(m), "bad %q got %d", s, len(m))
		if len(suggs) > 0 {
			require.Equal(t, suggs, m[0].GetSuggestedReplacements(), "sugg %q", s)
		}
	}

	assert0("як-небудь")
	assert0("А\u2013Т")
	assert0("ХХ\u2013ХХІ")

	assert1("яскраво\u2013рожевий.", "яскраво-рожевий", "яскраво \u2014 рожевий")
	assert1("яскраво\u2013шуруровий.", "яскраво-шуруровий", "яскраво \u2014 шуруровий")
	assert0("ХХ\u2014ХХІ")
	assert0("Вовка,— волкова")

	assert1("Вовка\u2014Волкова.", "Вовка-Волкова", "Вовка \u2014 Волкова")
	assert1("цукерок —знову низька", "цукерок-знову", "цукерок \u2014 знову")
	assert1("—знову низька", "\u2014 знову")
	assert1("знову—", "знову \u2014")
	assert0("\u2014 Київ, 1994")
	assert0("\u2013 Київ, 1994")
	assert1("важливіше \u2013потенційні", "важливіше-потенційні", "важливіше \u2014 потенційні")
	assert0("Рахунки 1 класу –")
	assert0("\u2013")
	assert0(" \u2013")

	m := rule.Match(languagetool.AnalyzePlain("любили ,—люби"))
	require.Equal(t, 1, len(m))
	require.Equal(t, 1, len(m[0].GetSuggestedReplacements()))
}
