package tools

// Twin of languagetool-core/src/test/java/org/languagetool/tools/ToolsTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of ToolsTest.testCorrectTextFromMatches
func TestTools_languagetool_core_CorrectTextFromMatches(t *testing.T) {
	matches := []TextMatch{
		{FromPos: 0, ToPos: 9, SuggestedReplacements: []string{"I've had"}},
		{FromPos: 0, ToPos: 9, SuggestedReplacements: []string{"I have"}},
	}
	require.Equal(t, "I've had", CorrectTextFromMatches("I've have", matches))
}

// UTF-16 positions: "café X" — é is one UTF-16 unit (U+00E9), not two UTF-8 bytes.
func TestTools_CorrectTextFromMatches_UTF16(t *testing.T) {
	// "café X" UTF-16 units: c a f é space X → positions 0..5; "X" at 5..6
	src := "café X"
	matches := []TextMatch{
		{FromPos: 5, ToPos: 6, SuggestedReplacements: []string{"Y"}},
	}
	require.Equal(t, "café Y", CorrectTextFromMatches(src, matches))
	// already-corrected skip: two overlapping matches same span, first wins
	matches2 := []TextMatch{
		{FromPos: 5, ToPos: 6, SuggestedReplacements: []string{"Y"}},
		{FromPos: 5, ToPos: 6, SuggestedReplacements: []string{"Z"}},
	}
	require.Equal(t, "café Y", CorrectTextFromMatches(src, matches2))
}

// Port of ToolsTest.testSelectRules
func TestTools_languagetool_core_SelectRules(t *testing.T) {
	newDemo := func() *RuleSelector {
		rs := NewRuleSelector("DEMO_RULE", "OTHER_RULE")
		rs.SetCategory("DEMO_RULE", "MISC")
		rs.SetCategory("OTHER_RULE", "CASING")
		return rs
	}
	set := func(ids ...string) map[string]struct{} {
		m := map[string]struct{}{}
		for _, id := range ids {
			m[id] = struct{}{}
		}
		return m
	}

	// empty → DEMO_RULE active
	rs := newDemo()
	rs.SelectRules(nil, nil, nil, nil, false, false)
	require.True(t, rs.Has("DEMO_RULE"))

	// disable DEMO_RULE by id
	rs = newDemo()
	rs.SelectRules(nil, nil, set("DEMO_RULE"), nil, false, false)
	require.False(t, rs.Has("DEMO_RULE"))

	// disable MISC category
	rs = newDemo()
	rs.SelectRules(set("MISC"), nil, nil, nil, false, false)
	require.False(t, rs.Has("DEMO_RULE"))

	// disable category but enable rule
	rs = newDemo()
	rs.SelectRules(set("MISC"), nil, nil, set("DEMO_RULE"), false, false)
	require.True(t, rs.Has("DEMO_RULE"))

	// useEnabledOnly=true, empty → still keeps default-on DEMO_RULE
	rs = newDemo()
	rs.SelectRules(nil, nil, nil, nil, true, false)
	require.True(t, rs.Has("DEMO_RULE"))

	// useEnabledOnly disable rule
	rs = newDemo()
	rs.SelectRules(nil, nil, set("DEMO_RULE"), nil, true, false)
	require.False(t, rs.Has("DEMO_RULE"))

	// useEnabledOnly disable category
	rs = newDemo()
	rs.SelectRules(set("MISC"), nil, nil, nil, true, false)
	require.False(t, rs.Has("DEMO_RULE"))

	// useEnabledOnly enable MISC category only
	rs = newDemo()
	rs.SelectRules(nil, set("MISC"), nil, nil, true, false)
	require.True(t, rs.Has("DEMO_RULE"))
	require.False(t, rs.Has("OTHER_RULE"))

	// useEnabledOnly enable CASING only → no DEMO_RULE
	rs = newDemo()
	rs.SelectRules(nil, set("CASING"), nil, nil, true, false)
	require.False(t, rs.Has("DEMO_RULE"))
	require.True(t, rs.Has("OTHER_RULE"))

	// enableTempOff activates default temp_off rules (Java Tools.selectRules)
	rs = newDemo()
	rs.SetDefaultTempOff("TEMP_OFF_RULE")
	require.False(t, rs.Has("TEMP_OFF_RULE"))
	rs.SelectRules(nil, nil, nil, nil, false, true)
	require.True(t, rs.Has("TEMP_OFF_RULE"))
}
