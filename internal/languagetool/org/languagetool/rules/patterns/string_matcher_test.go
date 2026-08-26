package patterns

import (
	"strings"
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

func TestStringMatcher_RequiredSubstringsMatches(t *testing.T) {
	// Java create path: required prefilter + checkCanReplaceRegex exhaustive.
	// PRP.+ → exhaustive [PRP) with minLength+1 for +
	m := NewStringMatcherRegexp("PRP.+")
	require.True(t, m.substringsSufficient)
	require.NotNil(t, m.required)
	require.Nil(t, m.re)
	require.True(t, m.Matches("PRPfoo"))
	require.True(t, m.Matches("PRPx"))
	require.False(t, m.Matches("PRP")) // + needs at least one more char
	require.False(t, m.Matches("XPRPfoo"))

	// foo.*bar exhaustive when MustStart+MustEnd
	m2 := NewStringMatcherRegexp("foo.*bar")
	require.True(t, m2.substringsSufficient)
	require.True(t, m2.Matches("foobar"))
	require.True(t, m2.Matches("fooXbar"))
	require.False(t, m2.Matches("foXbar"))
	require.False(t, m2.Matches("fooXba"))

	// unbounded that is not substring-sufficient still uses regex
	m3 := NewStringMatcherRegexp("x.+y")
	// a.+b form is sufficient; x.+y same
	require.True(t, m3.Matches("xay"))
	require.False(t, m3.Matches("xy")) // + needs content between? x.+y means x, one+, y — "xy" fails
	require.False(t, m3.Matches("xaz"))

	// MaxMatchLength uses UTF-16 length (Java s.length())
	long := strings.Repeat("a", MaxMatchLength+1)
	require.False(t, NewStringMatcher("a+", true, true).Matches(long))
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
