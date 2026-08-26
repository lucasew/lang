package xx

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDemoDisambiguator2Identity(t *testing.T) {
	d := NewDemoDisambiguator2()
	tok := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("x", nil, nil))
	s := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{tok})
	require.Equal(t, s.GetText(), d.Disambiguate(s).GetText())
}
