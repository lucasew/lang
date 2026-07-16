package tools

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuleMatchesJSONSerializerPort(t *testing.T) {
	s := NewRuleMatchesAsJsonSerializer()
	s.LanguageCode = "en-US"
	s.LanguageName = "English (US)"
	m := MatchForJSON{
		Message:               "bad <suggestion>good</suggestion>",
		FromPos:               5,
		ToPos:                 10,
		SuggestedReplacements: []string{"good"},
		RuleID:                "DEMO",
		RuleDescription:       "demo rule",
	}
	out, err := s.RuleMatchesToJSON([]MatchForJSON{m}, "hello world", 2)
	require.NoError(t, err)
	var resp ResponseJSON
	require.NoError(t, json.Unmarshal([]byte(out), &resp))
	require.Len(t, resp.Matches, 1)
	require.Equal(t, 5, resp.Matches[0].Offset)
	require.Equal(t, 5, resp.Matches[0].Length)
	require.Equal(t, "DEMO", resp.Matches[0].Rule.ID)
	require.Equal(t, "bad good", resp.Matches[0].Message)
	require.Equal(t, "good", resp.Matches[0].Replacements[0].Value)
	require.NotNil(t, resp.Software)
}
