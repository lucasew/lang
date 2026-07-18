package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSoftPostag_NumOrdNotDigitOnly(t *testing.T) {
	require.False(t, softPostagIsNumberOnly("Num:Ord"))
	require.False(t, softPostagIsNumberOnly("Num:Dig:Ord"))
	require.True(t, softPostagIsNumberOnly("Z.+"))
	require.True(t, softPostagIsNumberOnly("CD"))

	pt := NewPatternToken("", false, false, false)
	pt.SetPosToken(PosToken{PosTag: "Num:Ord", Regexp: false})
	m := NewPatternTokenMatcher(pt)
	// spelled ordinal should soft-match as open-class word, not require digits
	tok := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("tríú", nil, nil))
	require.True(t, m.IsMatchedReadings(tok), "tríú soft-matches Num:Ord")
}
