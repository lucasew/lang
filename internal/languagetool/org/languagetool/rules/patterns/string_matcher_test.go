package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringMatcherLiteralPort(t *testing.T) {
	m := NewStringMatcher("hello", false, true)
	require.True(t, m.Matches("hello"))
	require.False(t, m.Matches("Hello"))
	m2 := NewStringMatcher("hello", false, false)
	require.True(t, m2.Matches("Hello"))
}

func TestStringMatcherAlternationPort(t *testing.T) {
	m := NewStringMatcherRegexp("aa|bb")
	require.True(t, m.Matches("aa"))
	require.True(t, m.Matches("bb"))
	require.False(t, m.Matches("cc"))
	vals := m.GetPossibleValues()
	require.Contains(t, vals, "aa")
	require.Contains(t, vals, "bb")
}

func TestStringMatcherPossibleValuesPort(t *testing.T) {
	require.Nil(t, GetPossibleRegexpValues("x.*"))
	require.Nil(t, GetPossibleRegexpValues("a.c"))
	vals := GetPossibleRegexpValues("aa|bb")
	require.Contains(t, vals, "aa")
	require.Contains(t, vals, "bb")
	vals = GetPossibleRegexpValues("a(b|c)")
	require.Contains(t, vals, "ab")
	require.Contains(t, vals, "ac")
	vals = GetPossibleRegexpValues("[abc]")
	require.Len(t, vals, 3)
	vals = GetPossibleRegexpValues("are|is|w(?:as|ere)")
	require.Contains(t, vals, "was")
	require.Contains(t, vals, "were")
}

func TestStringMatcherMaxLengthPort(t *testing.T) {
	m := NewStringMatcher("x", false, true)
	long := make([]byte, MaxMatchLength+1)
	for i := range long {
		long[i] = 'x'
	}
	require.False(t, m.Matches(string(long)))
}

func TestStringMatcherInvalidSyntaxPort(t *testing.T) {
	require.Panics(t, func() { NewStringMatcherRegexp("tú|?") })
}

func TestStringMatcher_GetRequiredSubstrings(t *testing.T) {
	// Twin of StringMatcherTest.requiredSubstrings (selected cases).
	assertReq := func(re string, want *string) {
		t.Helper()
		got := GetRequiredSubstrings(re)
		if want == nil {
			require.Nil(t, got, "regexp=%q", re)
			return
		}
		require.NotNil(t, got, "regexp=%q", re)
		require.Equal(t, *want, got.String(), "regexp=%q", re)
	}
	s := func(v string) *string { return &v }
	assertReq("", s("[]"))
	assertReq("foo", s("[foo]"))
	assertReq("foo|bar", nil)
	assertReq(`\w`, nil)
	assertReq("PRP.+", s("[PRP)"))
	assertReq(".*PRP.+", s("(PRP)"))
	assertReq(".*PRP", s("(PRP]"))
	assertReq(".+PRP", s("(PRP]"))
	assertReq("a.+b", s("[a, b]"))
	assertReq("a.*b", s("[a, b]"))
	assertReq(`\bZünglein an der (Wage)\b`, s("[Zünglein an der Wage]"))
	assertReq(`(ökumenische[rn]?) (.*Messen?)`, s("[ökumenische,  , Messe)"))
	assertReq(`(CO2|Kohlendioxid|Schadstoff)\-?Emulsion(en)?`, s("(Emulsion)"))
	assertReq(`\bder (\w*(Verkehrs|Verbots|Namens|Hinweis|Warn)schild)`, s("[der , schild]"))
	assertReq(`\bvon Seiten\b`, s("[von Seiten]"))
	assertReq("((\\-)?[0-9]+[0-9.,]{0,15})(?:[\\s \u00a0\u202f]+)(°[^CFK])", s("(°)"))
	assertReq(`\b(teils\s[^,]+\steils)\b`, s("[teils, teils]"))
	assertReq(`§ ?(\d+[a-z]?)`, s("[§)"))
}
