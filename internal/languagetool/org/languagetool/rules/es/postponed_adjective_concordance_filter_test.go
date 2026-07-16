package es

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestPostponedAdjectiveConcordanceFilter_Disagrees(t *testing.T) {
	f := NewPostponedAdjectiveConcordanceFilter()
	nom := "NCFS000"
	adj := "AQ0MS0"
	tokens := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("casa", &nom, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("rojo", &adj, nil)),
	}
	match := rules.NewRuleMatch(f, nil, 0, 9, "msg")
	require.NotNil(t, f.AcceptRuleMatch(match, nil, tokens))
}

func TestPostponedAdjectiveConcordanceFilter_Agrees(t *testing.T) {
	f := NewPostponedAdjectiveConcordanceFilter()
	nom := "NCFS000"
	adj := "AQ0FS0"
	tokens := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("casa", &nom, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("roja", &adj, nil)),
	}
	match := rules.NewRuleMatch(f, nil, 0, 9, "msg")
	require.Nil(t, f.AcceptRuleMatch(match, nil, tokens))
}

func TestPostponedAdjectiveConcordanceFilter_NoAdjTag(t *testing.T) {
	f := NewPostponedAdjectiveConcordanceFilter()
	tokens := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("casa", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("x", nil, nil)),
	}
	match := rules.NewRuleMatch(f, nil, 0, 6, "msg")
	require.NotNil(t, f.AcceptRuleMatch(match, nil, tokens))
}
