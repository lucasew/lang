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

func TestTokenAgreementPrepNounRule(t *testing.T) {
	r := NewTokenAgreementPrepNounRule()
	require.Equal(t, TokenAgreementPrepNounRuleID, r.GetID())
	// "в" governs v_zna / v_mis / v_rod — instrumental (v_oru) is wrong.
	prepLemma := "в"
	sentBad := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("в", &prepLemma, "prep"),
		atr("домі", "noun:inanim:m:v_oru"),
	})
	require.NotEmpty(t, r.Match(sentBad), "prep+wrong case should match")

	sentGood := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("в", &prepLemma, "prep"),
		atr("домі", "noun:inanim:m:v_mis"),
	})
	require.Empty(t, r.Match(sentGood), "prep+governed case should pass")
}

func atrLemma(token string, lemma *string, tags ...string) *languagetool.AnalyzedTokenReadings {
	var readings []*languagetool.AnalyzedToken
	for _, tg := range tags {
		t := tg
		readings = append(readings, languagetool.NewAnalyzedToken(token, &t, lemma))
	}
	if len(readings) == 0 {
		readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(token, nil, lemma)}
	}
	return languagetool.NewAnalyzedTokenReadingsList(readings, 0)
}

func TestNounVerbOverlap(t *testing.T) {
	// person 3 singular both
	require.True(t, VerbInflectionsOverlap(
		[]string{"verb:m:3"},
		[]string{"noun:anim:m:v_naz"},
	))
}

func TestVerbNounCaseAgreeHelper(t *testing.T) {
	cg := LoadCaseGovernmentHelper()
	vLemma := "боятися"
	verb := atrLemma("боятися", &vLemma, "verb:imperf:inf")
	noun := atr("закордоном", "noun:inanim:m:v_oru")
	require.False(t, VerbNounCaseAgree(cg, verb, noun))
	noun2 := atr("закордону", "noun:inanim:m:v_rod")
	require.True(t, VerbNounCaseAgree(cg, verb, noun2))
}

func TestHasVidmPosTag(t *testing.T) {
	// noun has v_oru among wanted
	require.True(t, HasVidmPosTag([]string{"v_oru"}, atr("домі", "noun:inanim:m:v_oru")))
	require.False(t, HasVidmPosTag([]string{"v_zna"}, atr("домі", "noun:inanim:m:v_oru")))
	// :nv short-circuit
	require.True(t, HasVidmPosTag([]string{"v_oru"}, atr("щось", "noun:inanim:n:v_naz:nv")))
}

func TestCaseGovernmentDerivativesMerged(t *testing.T) {
	cg := LoadCaseGovernmentHelper()
	// static Java override
	require.True(t, cg.HasCaseGovernment("згідно з", "v_oru"))
}

func TestAdjNounException_EarlyArms(t *testing.T) {
	// голому сорочка
	sent := []*languagetool.AnalyzedTokenReadings{
		atr("X"), // pad index 0 if needed
		atr("голому", "adj:m:v_dav"),
		atr("сорочка", "noun:inanim:f:v_naz"),
	}
	// tokens as sentence without SENT_START - positions 0,1
	require.True(t, IsAdjNounException(sent[1:], 0, 1))

	// перший + ordinary noun (not FakeFem inanim:m)
	lem := "перший"
	tokens := []*languagetool.AnalyzedTokenReadings{
		atr("pad"),
		atrLemma("перший", &lem, "adj:m:v_naz:numr"),
		atr("голодування", "noun:inanim:n:v_zna"),
	}
	require.True(t, IsAdjNounException(tokens, 1, 2))
}

func TestAdjNounException_MoreArms(t *testing.T) {
	// на повну
	tokens := []*languagetool.AnalyzedTokenReadings{
		atr("X"),
		atr("на", "prep"),
		atr("повну", "adj:f:v_zna"),
		atr("людей", "noun:anim:p:v_rod"),
	}
	require.True(t, IsAdjNounException(tokens, 2, 3))

	// здатний
	lem := "здатний"
	tokens2 := []*languagetool.AnalyzedTokenReadings{
		atrLemma("здатні", &lem, "adj:p:v_naz"),
		atr("екскаватором", "noun:inanim:m:v_oru"),
	}
	require.True(t, IsAdjNounException(tokens2, 0, 1))

	// вольному воля
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("вольному", "adj:m:v_dav"),
		atr("воля", "noun:inanim:f:v_naz"),
	}, 0, 1))
}

func TestPrepNounException_EarlyArms(t *testing.T) {
	// на + capitalized v_rod
	prep := "на"
	name := "Купала"
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrLemma("на", &prep, "prep"),
		atrLemma("Купала", &name, "noun:anim:m:v_rod:prop:lname"),
	}
	require.True(t, IsPrepNounException(tokens, 0, 1))

	// при їх
	require.True(t, IsPrepNounException([]*languagetool.AnalyzedTokenReadings{
		atr("при", "prep"),
		atr("їх", "pron:pers:p:v_rod"),
	}, 0, 1))

	// normal pair not exception by these arms alone
	require.False(t, IsPrepNounException([]*languagetool.AnalyzedTokenReadings{
		atr("в", "prep"),
		atr("домі", "noun:inanim:m:v_mis"),
	}, 0, 1))
}

func TestPrepNounException_VidDoPlusMinus(t *testing.T) {
	require.True(t, IsPrepNounException([]*languagetool.AnalyzedTokenReadings{
		atr("від", "prep"), atr("а", "part"),
	}, 0, 1))
	require.True(t, IsPrepNounException([]*languagetool.AnalyzedTokenReadings{
		atr("до", "prep"), atr("я", "noun"),
	}, 0, 1))
	// мінус 1
	require.True(t, IsPrepNounException([]*languagetool.AnalyzedTokenReadings{
		atr("від", "prep"), atr("мінус", "noun"), atr("1", "num"),
	}, 0, 1))
}

func TestNumrNounException_SurfaceArms(t *testing.T) {
	require.True(t, IsNumrNounException([]*languagetool.AnalyzedTokenReadings{
		atr("багатьох", "numr:p:v_rod"), atr("людей", "noun:anim:p:v_rod"),
	}, 0, 1))
	require.True(t, IsNumrNounException([]*languagetool.AnalyzedTokenReadings{
		atr("дві", "numr:f:v_naz"), atr("ранку", "noun:inanim:m:v_rod"),
	}, 0, 1))
	// ordinary pair not covered by these surface arms
	require.False(t, IsNumrNounException([]*languagetool.AnalyzedTokenReadings{
		atr("дві", "numr:f:v_naz"), atr("книги", "noun:inanim:f:v_naz"),
	}, 0, 1))
}
