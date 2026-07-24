package xx

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDemoDisambiguator(t *testing.T) {
	d := NewDemoDisambiguator()
	require.Nil(t, d.Disambiguate(nil))
	tok := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("x", nil, nil))
	s := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{tok})
	require.Equal(t, s, d.Disambiguate(s))
	require.Equal(t, s, d.PreDisambiguate(s))
}
