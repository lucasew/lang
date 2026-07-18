package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSoftPostagNumberOnly_Z(t *testing.T) {
	require.True(t, softPostagIsNumberOnly("Z.+"))
	require.True(t, softPostagIsNumberOnly("Z"))
	require.True(t, softLooksLikeNumber("12"))
	require.True(t, softLooksLikeNumber("1.000"))
	require.False(t, softLooksLikeNumber("É"))
	require.False(t, softLooksLikeNumber("preciso"))

	// postag-only Z.+ must match digits, not words
	pt := NewPatternToken("", false, false, false)
	pt.SetPosToken(PosToken{PosTag: "Z.+", Regexp: true})
	m := NewPatternTokenMatcher(pt)

	word := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("É", nil, nil))
	require.False(t, m.IsMatchedReadings(word), "letter must not soft-match Z.+")

	num := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("12", nil, nil))
	require.True(t, m.IsMatchedReadings(num), "digits soft-match Z.+")
}
