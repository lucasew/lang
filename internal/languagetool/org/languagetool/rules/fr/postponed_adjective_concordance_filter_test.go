package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestPostponedAdjectiveConcordanceFilter_Disagrees(t *testing.T) {
	f := NewPostponedAdjectiveConcordanceFilter()
	nom := "N f s"
	adj := "J m s"
	tokens := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("maison", &nom, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("beau", &adj, nil)),
	}
	match := rules.NewRuleMatch(f, nil, 0, 12, "msg")
	require.NotNil(t, f.AcceptRuleMatch(match, nil, tokens))
}

func TestPostponedAdjectiveConcordanceFilter_Agrees(t *testing.T) {
	f := NewPostponedAdjectiveConcordanceFilter()
	nom := "N f s"
	adj := "J f s"
	tokens := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("maison", &nom, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("belle", &adj, nil)),
	}
	match := rules.NewRuleMatch(f, nil, 0, 12, "msg")
	require.Nil(t, f.AcceptRuleMatch(match, nil, tokens))
}
