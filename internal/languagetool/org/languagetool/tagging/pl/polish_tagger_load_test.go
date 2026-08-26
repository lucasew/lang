package pl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsureDefaultPolishTagger_POS(t *testing.T) {
	if DiscoverPolishPOSDict() == "" {
		t.Skip("polish.dict not in tree (inspiration pl module)")
	}
	EnsureDefaultPolishTagger()
	require.NotEmpty(t, PolishPOSDictPath())
	require.NotNil(t, DefaultPolishTagger)

	// Spot-check tags used by PolishWordTokenizer hybrid hyphen logic
	atrs := DefaultPolishTagger.Tag([]string{"polsko", "kobieta", "osiemnaście", "SMS-y"})
	require.Len(t, atrs, 4)
	require.True(t, atrs[0].IsTagged())
	require.True(t, atrs[0].HasPosTag("adja"))
	require.True(t, atrs[1].HasPartialPosTag("subst:"))
	require.True(t, atrs[2].HasPartialPosTag("num:"))
	// morphological ending form stays tagged (not split as compound)
	require.True(t, atrs[3].IsTagged())
}
