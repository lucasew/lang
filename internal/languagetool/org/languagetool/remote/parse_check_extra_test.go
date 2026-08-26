package remote

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseCheckJSON_TypeAndSentenceRanges(t *testing.T) {
	raw := `{
	  "software":{"name":"LanguageTool-Go","version":"dev","buildDate":"dev"},
	  "language":{"name":"English","code":"en"},
	  "matches":[{
	    "message":"Use a",
	    "shortMessage":"Wrong article",
	    "offset":8,"length":2,
	    "contextForSureMatch":0,
	    "type":{"typeName":"grammar"},
	    "context":{"text":"This is an test.","offset":8,"length":2},
	    "replacements":[{"value":"a"}],
	    "rule":{"id":"EN_A_VS_AN","description":"a/an","issueType":"grammar",
	      "category":{"id":"GRAMMAR","name":"Grammar"}}
	  }],
	  "sentenceRanges":[{"offset":0,"length":16}],
	  "ignoreRanges":[]
	}`
	res, err := ParseCheckJSON([]byte(raw))
	require.NoError(t, err)
	require.Len(t, res.Matches, 1)
	require.Equal(t, "grammar", res.Matches[0].GetTypeName())
	require.Equal(t, "EN_A_VS_AN", res.Matches[0].GetRuleID())
	require.Len(t, res.GetSentenceRanges(), 1)
	require.Equal(t, 0, res.GetSentenceRanges()[0].Offset)
	require.Equal(t, 16, res.GetSentenceRanges()[0].Length)
}
