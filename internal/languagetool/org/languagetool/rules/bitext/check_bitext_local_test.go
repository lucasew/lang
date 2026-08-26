package bitext

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckBitextLocal(t *testing.T) {
	// DifferentLengthRule / SameTranslation etc. may fire on contrived pairs
	ms := CheckBitextLocal("Hello world", "Hello")
	_ = ms
	// raw bitext then convert
	raw := CheckBitext("Same sentence.", "Same sentence.", nil)
	lm := ToLocalMatches(raw)
	require.Equal(t, len(raw), len(lm))
	for i := range lm {
		require.Equal(t, raw[i].RuleID, lm[i].RuleID)
		require.Equal(t, raw[i].FromPos, lm[i].FromPos)
	}
}
