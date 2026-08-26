package disambiguation

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRemovePreviousTags_ViaDisambiguate(t *testing.T) {
	// Java EnglishHybridDisambiguator: setRemovePreviousTags(true) turns
	// <NNP></NNP> chunk annotations into plain NNP readings.
	lines := []string{"New York\tNNP"}
	c := NewMultiWordChunker(lines, MultiWordChunkerSettings{
		AllowFirstCapitalized: true,
		AllowTitlecase:        true,
	})
	c.SetRemovePreviousTags(true)

	lt := languagetool.NewJLanguageTool("en")
	lt.TagWord = func(token string) []languagetool.TokenTag {
		switch token {
		case "New", "York":
			return []languagetool.TokenTag{{POS: "NNP", Lemma: token}}
		default:
			return nil
		}
	}
	sents := lt.Analyze("New York")
	require.NotEmpty(t, sents)
	out := c.Disambiguate(sents[0])
	require.NotNil(t, out)
	var sawNNP bool
	for _, tok := range out.GetTokensWithoutWhitespace() {
		for _, r := range tok.GetReadings() {
			if r.GetPOSTag() == nil {
				continue
			}
			require.NotContains(t, *r.GetPOSTag(), "<", "angle-bracket chunk tags should be flattened")
			if *r.GetPOSTag() == "NNP" {
				sawNNP = true
			}
		}
	}
	require.True(t, sawNNP, "expected plain NNP after removePreviousTags")
}
