package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func ptr(s string) *string { return &s }

func atr(token, pos, lemma string) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(token, ptr(pos), ptr(lemma)))
}

func sentenceWithoutWS(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	start := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("", ptr(languagetool.SentenceStartTagName), nil))
	all := append([]*languagetool.AnalyzedTokenReadings{start}, toks...)
	return languagetool.NewAnalyzedSentence(all)
}

func TestCAPostponedAdjectiveRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ca.PostponedAdjectiveConcordanceFilter"))
}

func TestPostponedAdjectiveConcordanceFilter_NoSentence(t *testing.T) {
	f := NewPostponedAdjectiveConcordanceFilter()
	require.Nil(t, f.AcceptRuleMatch(nil, nil, 0, nil, nil))
	require.Nil(t, f.AcceptRuleMatch(&rules.RuleMatch{}, nil, 0, nil, nil))
}

// "casa bonic" — FS noun + MS adj → suggest "bonica"
func TestPostponedAdjectiveConcordanceFilter_CasaBonic(t *testing.T) {
	f := NewPostponedAdjectiveConcordanceFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, posTagRegex string) []string {
		if posTagRegex == "A..FS.|V.P..SF.|PX.FS.*" {
			return []string{"bonica"}
		}
		return nil
	}
	sent := sentenceWithoutWS(
		atr("casa", "NCFS000", "casa"),
		atr("bonic", "AQ0MS0", "bonic"),
	)
	m := rules.NewRuleMatch(nil, sent, 5, 10, "msg")
	out := f.AcceptRuleMatch(m, nil, 2, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"bonica"}, out.GetSuggestedReplacements())
}

func TestPostponedAdjectiveConcordanceFilter_Agreeing(t *testing.T) {
	f := NewPostponedAdjectiveConcordanceFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, posTagRegex string) []string {
		return []string{"should-not-appear"}
	}
	sent := sentenceWithoutWS(
		atr("llibre", "NCMS000", "llibre"),
		atr("bonic", "AQ0MS0", "bonic"),
	)
	m := rules.NewRuleMatch(nil, sent, 7, 12, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, nil, 2, nil, nil))
}

func TestPostponedAdjectiveConcordanceFilter_NoSynth(t *testing.T) {
	f := NewPostponedAdjectiveConcordanceFilter()
	sent := sentenceWithoutWS(
		atr("casa", "NCFS000", "casa"),
		atr("bonic", "AQ0MS0", "bonic"),
	)
	m := rules.NewRuleMatch(nil, sent, 5, 10, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, nil, 2, nil, nil))
}

func TestCAFullMatch_UnknownNotNom(t *testing.T) {
	r := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("x", nil, nil))
	require.True(t, caMatchPostagRegexp(r, caKEEPCOUNT))
	require.False(t, caMatchPostagRegexp(r, caNOM))
}
