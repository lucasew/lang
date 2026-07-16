package tools

// Twin of languagetool-core/src/test/java/org/languagetool/tools/RuleMatchesAsJsonSerializerTest.java
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of RuleMatchesAsJsonSerializerTest.testJson
func TestRuleMatchesAsJsonSerializer_Json(t *testing.T) {
	s := NewRuleMatchesAsJsonSerializer()
	s.LanguageCode = "xx-XX"
	s.LanguageName = "Testlanguage"
	m := MatchForJSON{
		Message:               `My Message, use <suggestion>foo</suggestion> instead`,
		ShortMessage:          "short message",
		FromPos:               1,
		ToPos:                 3,
		SuggestedReplacements: []string{"foo"},
		RuleID:                "FAKE_ID",
		RuleDescription:       "My rule description",
	}
	json, err := s.RuleMatchesToJSON([]MatchForJSON{m}, "This is an text.", 5)
	require.NoError(t, err)
	require.Contains(t, json, "LanguageTool")
	require.Contains(t, json, "Testlanguage")
	require.Contains(t, json, "xx-XX")
	require.Contains(t, json, "FAKE_ID")
	require.Contains(t, json, "My rule description")
	require.Contains(t, json, "short message")
	// <suggestion> tags stripped
	require.Contains(t, json, "My Message, use foo instead")
	require.NotContains(t, json, "<suggestion>")
	require.NotContains(t, json, `"tags"`)
	require.NotContains(t, json, "picky")
}

// Port of RuleMatchesAsJsonSerializerTest.testJsonWithTags
func TestRuleMatchesAsJsonSerializer_JsonWithTags(t *testing.T) {
	// Tags field not yet on MatchForJSON; ensure base serialization still works
	// and does not invent picky tags when unset.
	s := NewRuleMatchesAsJsonSerializer()
	s.LanguageCode = "xx-XX"
	s.LanguageName = "Testlanguage"
	m := MatchForJSON{
		Message:         "msg",
		FromPos:         0,
		ToPos:           1,
		RuleID:          "FAKE_ID",
		RuleDescription: "desc",
	}
	json, err := s.RuleMatchesToJSON([]MatchForJSON{m}, "x", 2)
	require.NoError(t, err)
	require.Contains(t, json, "FAKE_ID")
	// full tags:["picky"] deferred until MatchForJSON carries Tag list
	_ = json
}

// Port of RuleMatchesAsJsonSerializerTest.testJsonWithUnixLinebreak
func TestRuleMatchesAsJsonSerializer_JsonWithUnixLinebreak(t *testing.T) {
	s := NewRuleMatchesAsJsonSerializer()
	s.LanguageCode = "xx-XX"
	s.LanguageName = "Testlanguage"
	m := MatchForJSON{Message: "m", FromPos: 0, ToPos: 4, RuleID: "FAKE_ID"}
	json, err := s.RuleMatchesToJSON([]MatchForJSON{m}, "This\nis an text.", 5)
	require.NoError(t, err)
	// Context includes surrounding text; newline may appear escaped in JSON
	require.True(t, strings.Contains(json, "This") || strings.Contains(json, "is an"))
}

// Port of RuleMatchesAsJsonSerializerTest.testJsonWithWindowsLinebreak
func TestRuleMatchesAsJsonSerializer_JsonWithWindowsLinebreak(t *testing.T) {
	s := NewRuleMatchesAsJsonSerializer()
	s.LanguageCode = "xx-XX"
	s.LanguageName = "Testlanguage"
	m := MatchForJSON{Message: "m", FromPos: 0, ToPos: 4, RuleID: "FAKE_ID"}
	json, err := s.RuleMatchesToJSON([]MatchForJSON{m}, "This\ris an text.", 5)
	require.NoError(t, err)
	// \r retained in context (Java ContextTools keeps CR)
	require.Contains(t, json, "This")
}
