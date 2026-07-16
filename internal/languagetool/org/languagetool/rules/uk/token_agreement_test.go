package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func atr(token string, tags ...string) *languagetool.AnalyzedTokenReadings {
	var readings []*languagetool.AnalyzedToken
	for _, tg := range tags {
		t := tg
		readings = append(readings, languagetool.NewAnalyzedToken(token, &t, nil))
	}
	if len(readings) == 0 {
		readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(token, nil, nil)}
	}
	return languagetool.NewAnalyzedTokenReadingsList(readings, 0)
}

func TestInflectionExtractAndAgree(t *testing.T) {
	adj := GetAdjCaseInflections([]string{"adj:m:v_naz:rinanim"})
	require.NotEmpty(t, adj)
	require.Equal(t, "m", adj[0].Gender)
	require.Equal(t, "v_naz", adj[0].Case)

	noun := GetNounCaseInflections([]string{"noun:inanim:m:v_naz"})
	require.NotEmpty(t, noun)
	require.True(t, AdjNounAgree(
		[]string{"adj:m:v_naz"},
		[]string{"noun:inanim:m:v_naz"},
	))
	require.False(t, AdjNounAgree(
		[]string{"adj:f:v_naz"},
		[]string{"noun:inanim:m:v_naz"},
	))
}

func TestTokenAgreementAdjNounRule(t *testing.T) {
	r := NewTokenAgreementAdjNounRule()
	require.Equal(t, TokenAgreementAdjNounRuleID, r.GetID())

	// disagreeing pair
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("велика", "adj:f:v_naz"),
		atr("будинок", "noun:inanim:m:v_naz"),
	})
	matches := r.Match(sent)
	require.NotEmpty(t, matches)

	// agreeing pair
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("великий", "adj:m:v_naz"),
		atr("будинок", "noun:inanim:m:v_naz"),
	})
	require.Empty(t, r.Match(sent2))
}

func TestTokenAgreementRulesConstruct(t *testing.T) {
	require.Equal(t, TokenAgreementNumrNounRuleID, NewTokenAgreementNumrNounRule().GetID())
	require.Equal(t, TokenAgreementPrepNounRuleID, NewTokenAgreementPrepNounRule().GetID())
	require.Equal(t, TokenAgreementNounVerbRuleID, NewTokenAgreementNounVerbRule().GetID())
	require.Equal(t, TokenAgreementVerbNounRuleID, NewTokenAgreementVerbNounRule().GetID())
}

func TestNounVerbOverlap(t *testing.T) {
	// person 3 singular both
	require.True(t, VerbInflectionsOverlap(
		[]string{"verb:m:3"},
		[]string{"noun:anim:m:v_naz"},
	))
}
