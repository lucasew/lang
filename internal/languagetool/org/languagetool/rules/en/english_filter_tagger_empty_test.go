package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestFilterTagWordToATR_EmptyDictHit(t *testing.T) {
	// Dict miss → untagged ATR (IsTagged false), never panic on empty readings.
	tw := func(token string) []languagetool.TokenTag { return nil }
	atr := filterTagWordToATR("xyzzy_not_a_word", tw)
	require.NotNil(t, atr)
	require.Equal(t, "xyzzy_not_a_word", atr.GetToken())
	require.False(t, atr.IsTagged())

	// Empty surface (SENT_START probe) also safe.
	atr2 := filterTagWordToATR("", tw)
	require.NotNil(t, atr2)
	require.False(t, atr2.IsTagged())
}
