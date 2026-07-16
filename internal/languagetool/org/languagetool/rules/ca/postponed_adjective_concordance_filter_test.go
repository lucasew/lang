package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestPostponedAdjectiveConcordanceFilter_Disagrees(t *testing.T) {
	f := NewPostponedAdjectiveConcordanceFilter()
	// casa (N.FS) + gran (A...MS) → disagree → keep match
	nom := "NCFS000"
	adj := "AQ0MS0"
	tokens := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("casa", &nom, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("gran", &adj, nil)),
	}
	match := rules.NewRuleMatch(f, nil, 0, 9, "msg")
	got := f.AcceptRuleMatch(match, nil, tokens)
	require.NotNil(t, got)
}

func TestPostponedAdjectiveConcordanceFilter_Agrees(t *testing.T) {
	f := NewPostponedAdjectiveConcordanceFilter()
	nom := "NCFS000"
	adj := "AQ0FS0"
	tokens := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("casa", &nom, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("bona", &adj, nil)),
	}
	match := rules.NewRuleMatch(f, nil, 0, 9, "msg")
	got := f.AcceptRuleMatch(match, nil, tokens)
	require.Nil(t, got)
}

func TestPostponedAdjectiveConcordanceFilter_AdjPosArg(t *testing.T) {
	f := NewPostponedAdjectiveConcordanceFilter()
	nom := "NCMS000"
	adj := "AQ0FS0"
	tokens := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("el", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("cotxe", &nom, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("vermella", &adj, nil)),
	}
	match := rules.NewRuleMatch(f, nil, 0, 20, "msg")
	got := f.AcceptRuleMatch(match, map[string]string{"adj_pos": "3"}, tokens)
	require.NotNil(t, got)
}

func TestGNAgrees(t *testing.T) {
	require.True(t, gn{m: true, s: true}.agrees(gn{m: true, s: true}))
	require.False(t, gn{m: true, s: true}.agrees(gn{f: true, s: true}))
	require.True(t, gn{}.agrees(gn{m: true, s: true}))
}
