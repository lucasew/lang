package languagetool

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestTagWordFromMap(t *testing.T) {
	m := tagging.MapWordTagger{
		"cats": {tagging.NewTaggedWord("cat", "NNS")},
	}
	lt := NewJLanguageTool("en")
	lt.TagWord = TagWordFromMap(m)
	sents := lt.Analyze("cats")
	require.NotEmpty(t, sents)
	found := false
	for _, tok := range sents[0].GetTokensWithoutWhitespace() {
		if tok.GetToken() == "cats" {
			require.Equal(t, "NNS", *tok.GetReadings()[0].GetPOSTag())
			require.Equal(t, "cat", *tok.GetReadings()[0].GetLemma())
			found = true
		}
	}
	require.True(t, found)
}
