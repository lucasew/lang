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

func TestNounVerbException_EarlyArms(t *testing.T) {
	// правда + verb
	require.True(t, IsNounVerbException([]*languagetool.AnalyzedTokenReadings{
		atr("правда", "noun:inanim:f:v_naz"),
		atr("було", "verb:imperf:past:n"),
	}, 0, 1))
	// під три чорти — Java mBefore needs room for iCond (SENT_START pad like LT tokens[0])
	require.True(t, IsNounVerbException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT_START"),
		atr("під", "prep"), atr("три", "numr"), atr("чорти", "noun:anim:p:v_naz"),
		atr("йдуть", "verb:imperf:pres:p:3"),
	}, 3, 4))
	// normal disagree pair not exception
	require.False(t, IsNounVerbException([]*languagetool.AnalyzedTokenReadings{
		atr("хлопець", "noun:anim:m:v_naz"),
		atr("читає", "verb:imperf:pres:s:3"),
	}, 0, 1))
}

func TestVerbNounException_EarlyArms(t *testing.T) {
	// хотіти + v_oru
	lem := "хотіти"
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atrLemma("хоче", &lem, "verb:imperf:pres:s:3"),
		atr("маляром", "noun:anim:m:v_oru"),
	}, 0, 1))
	// чим могла
	mog := "могти"
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("чим", "pron"), atrLemma("могла", &mog, "verb:imperf:past:f"),
		atr("силою", "noun:inanim:f:v_oru"),
	}, 1, 2))
}

func TestNounVerbException_MoreArms(t *testing.T) {
	// both capitalized
	require.True(t, IsCapitalized("Андрій") && IsCapitalized("Качала"))
	require.True(t, IsNounVerbException([]*languagetool.AnalyzedTokenReadings{
		atr("Андрій", "noun:anim:m:v_naz:prop:fname"),
		atr("Качала", "verb:imperf:past:m"),
	}, 0, 1))
	// all-upper verb
	require.True(t, IsNounVerbException([]*languagetool.AnalyzedTokenReadings{
		atr("Тарас", "noun:anim:m:v_naz:prop:fname"),
		atr("ЗАКУСИЛО", "verb:imperf:past:n"),
	}, 0, 1))
	// кандидат в президенти — prep lemma required for HasLemmaTokenAny
	prep := "в"
	require.True(t, IsNounVerbException([]*languagetool.AnalyzedTokenReadings{
		atr("кандидат", "noun"), atrLemma("в", &prep, "prep"),
		atr("президенти", "noun:anim:p:v_naz"),
		atr("поїхав", "verb:perf:past:m"),
	}, 2, 3))
}

func TestVerbNounException_MoreArms(t *testing.T) {
	// що є сил — Java requires verbPos > 1
	but := "бути"
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("от", "part"), atr("що", "conj"), atrLemma("є", &but, "verb:imperf:pres:s:3"),
		atr("сил", "noun:inanim:p:v_rod"),
	}, 2, 3))
	// був людина
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("був", "verb:imperf:past:m"),
		atr("людина", "noun:anim:f:v_naz"),
	}, 0, 1))
}

func TestNounVerbException_InfAgreementSearch(t *testing.T) {
	// громадяни вважатися after здатні — Java reverseSearch requires nounPos > 1
	zd := "здатний"
	require.True(t, IsNounVerbException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT"),
		atrLemma("здатні", &zd, "adj:p:v_naz"),
		atr("громадяни", "noun:anim:p:v_naz"),
		atr("вважатися", "verb:imperf:inf:refl"),
	}, 2, 3))
}

func TestVerbNounException_SearchHelperArms(t *testing.T) {
	// потрібно буде склянку
	treba := "потрібно"
	buti := "бути"
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atrLemma("потрібно", &treba, "adv"),
		atrLemma("буде", &buti, "verb:imperf:futr:s:3"),
		atr("склянку", "noun:inanim:f:v_zna"),
	}, 1, 2))

	// буде видно супутники
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("буде", "verb:imperf:futr:s:3"),
		atr("видно", "adv:predic"),
		atr("супутники", "noun:inanim:p:v_naz"),
	}, 0, 2))

	// став жовтого кольору
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("став", "verb:perf:past:m"),
		atr("жовтого", "adj:m:v_rod"),
		atr("кольору", "noun:inanim:m:v_rod"),
	}, 0, 1))

	// станом на
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("станом", "noun"), atr("на", "prep"), atr("1", "number"),
	}, 0, 2))
}
