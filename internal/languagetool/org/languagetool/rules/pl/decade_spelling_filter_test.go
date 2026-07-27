package pl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

// Outcome matrix derived from Java DecadeSpellingFilter source (no official Java unit test).

func TestDecadeSpellingFilter_GetRomanNumber(t *testing.T) {
	// Tables/algorithm twin of private getRomanNumber.
	require.Equal(t, "", getRomanNumber(0))
	require.Equal(t, "I", getRomanNumber(1))
	require.Equal(t, "IV", getRomanNumber(4))
	require.Equal(t, "V", getRomanNumber(5))
	require.Equal(t, "IX", getRomanNumber(9))
	require.Equal(t, "X", getRomanNumber(10))
	require.Equal(t, "XIV", getRomanNumber(14))
	require.Equal(t, "XIX", getRomanNumber(19))
	require.Equal(t, "XX", getRomanNumber(20))
	require.Equal(t, "XXI", getRomanNumber(21))
	require.Equal(t, "XL", getRomanNumber(40))
	require.Equal(t, "L", getRomanNumber(50))
	require.Equal(t, "XC", getRomanNumber(90))
	require.Equal(t, "C", getRomanNumber(100))
	require.Equal(t, "CD", getRomanNumber(400))
	require.Equal(t, "D", getRomanNumber(500))
	require.Equal(t, "CM", getRomanNumber(900))
	require.Equal(t, "M", getRomanNumber(1000))
	require.Equal(t, "MCMXCIX", getRomanNumber(1999))
	// Negative: loop never appends (Java same).
	require.Equal(t, "", getRomanNumber(-1))
}

func TestDecadeSpellingFilter_AcceptRuleMatch_Years(t *testing.T) {
	f := NewDecadeSpellingFilter()
	// 1990 → decade 90, century 19 → wiek XX (DATA_DEKADY example)
	out := acceptDecade(t, f, 3, 10, "lata {dekada}. wieku {wiek}", "1990")
	require.Equal(t, "lata 90. wieku XX", out.GetMessage())
	// 2000 → 00 / XXI
	out = acceptDecade(t, f, 0, 4, "{dekada}/{wiek}", "2000")
	require.Equal(t, "00/XXI", out.GetMessage())
	// 1840 → 40 / XIX (grammar.xml example)
	out = acceptDecade(t, f, 1, 6, "{dekada}. {wiek} w.", "1840")
	require.Equal(t, "40. XIX w.", out.GetMessage())
	// 1770 → 70 / XVIII
	out = acceptDecade(t, f, 1, 6, "{dekada}. {wiek} w.", "1770")
	require.Equal(t, "70. XVIII w.", out.GetMessage())
	// 1900 → 00 / XX (pierwsza dekada rule uses {wiek} only)
	out = acceptDecade(t, f, 1, 6, "pierwsza dekada {wiek} w.", "1900")
	require.Equal(t, "pierwsza dekada XX w.", out.GetMessage())
	// century "00" → cent+1 = 1 → I; decade empty when lata length == 2
	out = acceptDecade(t, f, 1, 2, "{dekada}|{wiek}", "00")
	require.Equal(t, "|I", out.GetMessage())
}

func TestDecadeSpellingFilter_AcceptRuleMatch_PreservesFields(t *testing.T) {
	f := NewDecadeSpellingFilter()
	rule := rules.NewFakeRule("DATA_DEKADY")
	m := rules.NewRuleMatch(rule, nil, 5, 12, "lata {dekada}. wieku {wiek}")
	m.SetPatternPosition(4, 13)
	m.ShortMessage = "Błędny zapis daty"
	m.SetType(rules.RuleMatchTypeHint)

	out := f.AcceptRuleMatch(m, map[string]string{"lata": "1990"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, "lata 90. wieku XX", out.GetMessage())
	require.Equal(t, 5, out.GetFromPos())
	require.Equal(t, 12, out.GetToPos())
	require.Equal(t, 4, out.GetPatternFromPos())
	require.Equal(t, 13, out.GetPatternToPos())
	require.Equal(t, "Błędny zapis daty", out.GetShortMessage())
	require.Equal(t, rules.RuleMatchTypeHint, out.GetType())
	require.Equal(t, rule, out.GetRule())
}

func TestDecadeSpellingFilter_AcceptRuleMatch_SuggestionsAndUppercase(t *testing.T) {
	f := NewDecadeSpellingFilter()
	// fromPos != 0 → no uppercaseFirstChar on suggestion body
	msg := `Błędnie. Poprawnie: <suggestion>{dekada}. {wiek} w.</suggestion>`
	out := acceptDecade(t, f, 3, 10, msg, "1990")
	require.Equal(t, []string{"90. XX w."}, out.GetSuggestedReplacements())

	// fromPos == 0 → startWithUppercase true (Java match.getFromPos() == 0)
	msg2 := `Błędnie. Poprawnie: <suggestion>pierwsza dekada {wiek} w.</suggestion>`
	out = acceptDecade(t, f, 0, 10, msg2, "1900")
	require.Equal(t, []string{"Pierwsza dekada XX w."}, out.GetSuggestedReplacements())

	// Digit-leading body: UppercaseFirstChar leaves digits unchanged
	msg3 := `x <suggestion>{dekada}. {wiek} w.</suggestion>`
	out = acceptDecade(t, f, 0, 5, msg3, "1840")
	require.Equal(t, []string{"40. XIX w."}, out.GetSuggestedReplacements())

	// <mistake/> bodies skipped (Java PatternRuleMatcher.MISTAKE)
	msg4 := `a <suggestion>keep {wiek}</suggestion> b <suggestion>bad<mistake/>x</suggestion>`
	out = acceptDecade(t, f, 1, 2, msg4, "2000")
	require.Equal(t, []string{"keep XXI"}, out.GetSuggestedReplacements())
}

func TestDecadeSpellingFilter_AcceptRuleMatch_ParseFailReturnsNil(t *testing.T) {
	f := NewDecadeSpellingFilter()
	rule := rules.NewFakeRule("DATA_DEKADY")
	m := rules.NewRuleMatch(rule, nil, 0, 4, "{dekada}/{wiek}")
	// Non-numeric century → NumberFormatException → IllegalArgumentException → null
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"lata": "ab90"}, 0, nil, nil))
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"lata": "xx"}, 0, nil, nil))
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"lata": "1x00"}, 0, nil, nil))
	// toPos <= fromPos → RuleMatch ctor IllegalArgumentException → null
	badPos := rules.NewRuleMatch(rule, nil, 5, 5, "{wiek}")
	// NewRuleMatch does not enforce to>from; AcceptRuleMatch must still mirror Java ctor check.
	badPos.FromPos, badPos.ToPos = 5, 5
	require.Nil(t, f.AcceptRuleMatch(badPos, map[string]string{"lata": "1990"}, 0, nil, nil))
	badPos.FromPos, badPos.ToPos = 6, 5
	require.Nil(t, f.AcceptRuleMatch(badPos, map[string]string{"lata": "1990"}, 0, nil, nil))
}

func TestDecadeSpellingFilter_AcceptRuleMatch_PropagatesNonIAE(t *testing.T) {
	f := NewDecadeSpellingFilter()
	rule := rules.NewFakeRule("DATA_DEKADY")
	m := rules.NewRuleMatch(rule, nil, 0, 4, "{dekada}/{wiek}")

	// Missing "lata" → Java NPE (not IAE)
	require.Panics(t, func() {
		f.AcceptRuleMatch(m, map[string]string{}, 0, nil, nil)
	})
	// nil arguments → Java NPE
	require.Panics(t, func() {
		f.AcceptRuleMatch(m, nil, 0, nil, nil)
	})
	// Short lata → StringIndexOutOfBoundsException (not IAE)
	require.Panics(t, func() {
		f.AcceptRuleMatch(m, map[string]string{"lata": ""}, 0, nil, nil)
	})
	require.Panics(t, func() {
		f.AcceptRuleMatch(m, map[string]string{"lata": "1"}, 0, nil, nil)
	})
	// null match after successful parse path still NPE on getMessage — order is substring first
	// so short-path: lata ok then match nil
	require.Panics(t, func() {
		f.AcceptRuleMatch(nil, map[string]string{"lata": "1990"}, 0, nil, nil)
	})
	// null rule → Objects.requireNonNull NPE
	mNilRule := rules.NewRuleMatch(nil, nil, 0, 4, "{wiek}")
	require.Panics(t, func() {
		f.AcceptRuleMatch(mNilRule, map[string]string{"lata": "1990"}, 0, nil, nil)
	})
}

func TestPLDecadeSpellingFilterRegistered(t *testing.T) {
	class := "org.languagetool.rules.pl.DecadeSpellingFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
	f := patterns.GlobalRuleFilterCreator.GetFilter(class)
	require.NotNil(t, f)
	m := rules.NewRuleMatch(rules.NewFakeRule("DATA_DEKADY"), nil, 0, 1, "{dekada}/{wiek}")
	out := f.AcceptRuleMatch(m, map[string]string{"lata": "2000"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, "00/XXI", out.GetMessage())
}

func acceptDecade(t *testing.T, f *DecadeSpellingFilter, from, to int, message, lata string) *rules.RuleMatch {
	t.Helper()
	m := rules.NewRuleMatch(rules.NewFakeRule("DATA_DEKADY"), nil, from, to, message)
	out := f.AcceptRuleMatch(m, map[string]string{"lata": lata}, 0, nil, nil)
	require.NotNil(t, out)
	return out
}
