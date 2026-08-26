package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUkrainianCommaWhitespaceException(t *testing.T) {
	r := NewUkrainianCommaWhitespaceRule(nil)
	require.NotNil(t, r.IsException)
	tokens := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
			languagetool.NewAnalyzedToken("a", nil, nil),
		}, 0),
		languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
			languagetool.NewAnalyzedToken("\u2014", nil, nil),
		}, 0),
	}
	require.True(t, r.IsException(tokens, 1))
}

func TestUkrainianUppercaseListException(t *testing.T) {
	r := NewUkrainianUppercaseSentenceStartRule(nil)
	start := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("", nil, nil),
	}, 0)
	tokens := []*languagetool.AnalyzedTokenReadings{
		start,
		languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
			languagetool.NewAnalyzedToken("а", nil, nil),
		}, 0),
		languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
			languagetool.NewAnalyzedToken(")", nil, nil),
		}, 0),
	}
	require.True(t, r.IsException(tokens, 1))
}
