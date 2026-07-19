package uk

import (
	"regexp"
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

func TestNounVerbException_GeoAndMore(t *testing.T) {
	// GEO: місто + Kyiv capitalized
	require.True(t, IsNounVerbException([]*languagetool.AnalyzedTokenReadings{
		atr("у", "prep"), atrLemma("місті", strPtr("місто"), "noun:inanim:n:v_mis"),
		atr("Київ", "noun:inanim:m:v_naz:prop"),
		atr("відбулося", "verb:perf:past:n"),
	}, 2, 3))

	// разом + p verb
	require.True(t, IsNounVerbException([]*languagetool.AnalyzedTokenReadings{
		atr("вони", "noun"), atr("разом", "adv"), atr("брали", "verb:imperf:past:p"),
	}, 0, 2))

	// ніж — Java requires nounPos > 2
	nizh := "ніж"
	require.True(t, IsNounVerbException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT"), atr("більше"), atrLemma("ніж", &nizh, "conj"), atr("будь-хто", "noun"),
		atr("маємо", "verb:imperf:pres:p:1"),
	}, 3, 4))

	// візьми та — lemma on conj
	ta := "та"
	require.True(t, IsNounVerbException([]*languagetool.AnalyzedTokenReadings{
		atr("вона", "noun"), atr("візьми", "verb"), atrLemma("та", &ta, "conj"), atr("скажи", "verb"),
	}, 0, 1))
}

func strPtr(s string) *string { return &s }

func TestLemmaHelper_TokenSearchAndProper(t *testing.T) {
	require.True(t, IsPossiblyProperNoun(atr("Київ")))
	require.True(t, IsInitial(atr("А.")))
	require.True(t, IsDash(atr("—")))
	tokens := []*languagetool.AnalyzedTokenReadings{
		atr("SENT"), atr("як"), atr("австрієць", "noun:anim:m:v_naz"),
	}
	require.Equal(t, 1, TokenSearch(tokens, 2, "", regexp.MustCompile(`^[Яя]к$`),
		regexp.MustCompile(`adj:.:v_naz.*`), DirReverse))
}

func TestNumrNounException_JavaArms(t *testing.T) {
	// багатьох — soft skip
	require.True(t, IsNumrNounException([]*languagetool.AnalyzedTokenReadings{
		atr("для", "prep"), atr("багатьох", "numr:p:v_rod"), atr("людей", "noun:anim:p:v_rod"),
	}, 1, 2))

	// noun surface ранок/ранку
	require.True(t, IsNumrNounException([]*languagetool.AnalyzedTokenReadings{
		atr("п'ять", "numr:p:v_naz"), atr("ранку", "noun:inanim:m:v_rod"),
	}, 0, 1))

	// lemma soft: весь
	v := "весь"
	require.True(t, IsNumrNounException([]*languagetool.AnalyzedTokenReadings{
		atr("три", "numr:p:v_naz"), atrLemma("весь", &v, "adj:m:v_naz"),
	}, 0, 1))

	// 22 червня
	ch := "червень"
	require.True(t, IsNumrNounException([]*languagetool.AnalyzedTokenReadings{
		atr("22", "number"), atrLemma("червня", &ch, "noun:inanim:m:v_rod"),
	}, 0, 1))

	// 3 / 4 — Java numrPos > 2 (SENT_START pad)
	require.True(t, IsNumrNounException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT_START"), atr("3", "number"), atr("/"), atr("4", "number"), atr("понеділка", "noun"),
	}, 3, 4))

	// № before — Java numrPos > 1
	require.True(t, IsNumrNounException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT_START"), atr("№"), atr("5", "number"), atr("розділ", "noun"),
	}, 2, 3))

	// adj.*numr
	require.True(t, IsNumrNounException([]*languagetool.AnalyzedTokenReadings{
		atr("двадцять", "numr"), atr("перший", "adj:m:v_naz:numr"),
	}, 0, 1))

	// обоє + anim p naz
	require.True(t, IsNumrNounException([]*languagetool.AnalyzedTokenReadings{
		atr("обоє", "numr"), atr("режисери", "noun:anim:p:v_naz"),
	}, 0, 1))

	// сьома вода
	require.True(t, IsNumrNounException([]*languagetool.AnalyzedTokenReadings{
		atr("сьома", "numr"), atr("вода", "noun:inanim:f:v_naz"),
	}, 0, 1))

	// два + adj p rod at end
	require.True(t, IsNumrNounException([]*languagetool.AnalyzedTokenReadings{
		atr("два", "numr:p:v_naz"), atr("нових", "adj:p:v_rod"),
	}, 0, 1))

	// хвилин п'ять — TIME_PLUS before numr; Java requires non-disjoint inflections
	hv := "хвилина"
	require.True(t, IsNumrNounException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT_START"),
		atrLemma("хвилин", &hv, "noun:inanim:p:v_rod"),
		atr("п'ять", "numr:p:v_rod"),
		atr("люди", "noun:anim:p:v_naz"),
	}, 2, 3))
}

func TestPrepNounException_StrongAndNonInfl(t *testing.T) {
	// до сьогодні
	require.True(t, IsPrepNounException([]*languagetool.AnalyzedTokenReadings{
		atr("до", "prep"), atr("сьогодні", "adv"),
	}, 0, 1))
	// на завтра
	require.True(t, IsPrepNounException([]*languagetool.AnalyzedTokenReadings{
		atr("на", "prep"), atr("завтра", "adv"),
	}, 0, 1))
	// за вчора
	require.True(t, IsPrepNounException([]*languagetool.AnalyzedTokenReadings{
		atr("за", "prep"), atr("вчора", "adv"),
	}, 0, 1))
	// у нікуди
	require.True(t, IsPrepNounException([]*languagetool.AnalyzedTokenReadings{
		atr("у", "prep"), atr("нікуди", "adv"),
	}, 0, 1))
	// не + adj
	require.True(t, IsPrepNounException([]*languagetool.AnalyzedTokenReadings{
		atr("до", "prep"), atr("не", "part"), atr("властиву", "adj:f:v_zna"),
	}, 0, 1))
	// ADV_QUANT
	chymalo := "чимало"
	require.True(t, IsPrepNounException([]*languagetool.AnalyzedTokenReadings{
		atr("про", "prep"), atrLemma("чимало", &chymalo, "adv"), atr("обмежень", "noun"),
	}, 0, 1))
	// Z_ZI_IZ + pseudo num
	sotnia := "сотня"
	require.True(t, IsPrepNounException([]*languagetool.AnalyzedTokenReadings{
		atr("із", "prep"), atrLemma("сотня", &sotnia, "noun:inanim:f:v_naz"),
	}, 0, 1))
	// part insert
	require.True(t, IsPrepNounException([]*languagetool.AnalyzedTokenReadings{
		atr("на", "prep"), atr("навіть", "part"), atr("день", "noun"),
	}, 0, 1))
	// лише
	require.True(t, IsPrepNounException([]*languagetool.AnalyzedTokenReadings{
		atr("на", "prep"), atr("лише", "part"), atr("день", "noun"),
	}, 0, 1))
	// наприклад
	require.True(t, IsPrepNounException([]*languagetool.AnalyzedTokenReadings{
		atr("в", "prep"), atr("наприклад", "insert"), atr("день", "noun"),
	}, 0, 1))
	// нічого не + adj
	require.True(t, IsPrepNounException([]*languagetool.AnalyzedTokenReadings{
		atr("на", "prep"), atr("нічого", "noun"), atr("не", "part"), atr("вартий", "adj:m:v_naz"),
	}, 0, 1))
	// в дев'яносто восьмому — numr with both naz and rod readings
	require.True(t, IsPrepNounException([]*languagetool.AnalyzedTokenReadings{
		atr("в", "prep"),
		atr("дев'яносто", "numr:p:v_naz", "numr:p:v_rod"),
		atr("восьмому", "numr:m:v_mis"),
	}, 0, 1))
	// замість inf
	require.True(t, IsPrepNounException([]*languagetool.AnalyzedTokenReadings{
		atr("замість", "prep"), atr("йому", "noun"), atr("засвоїти", "verb:imperf:inf"),
	}, 0, 1))
	// не те before — SENT_START pad (Java mBefore > 0)
	require.True(t, IsPrepNounException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT_START"), atr("не"), atr("те"), atr("по", "prep"), atr("лихим", "adj"),
	}, 3, 4))
}

func TestAdjNounException_MoreJavaArms(t *testing.T) {
	// гвардії + next noun overlap
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("молодшого", "adj:m:v_rod"), atr("гвардії", "noun:inanim:f:v_rod"),
		atr("сержанта", "noun:anim:m:v_rod"),
	}, 0, 1))

	// півстоліття
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("довгих", "adj:p:v_rod"), atr("півстоліття", "noun:inanim:n:v_rod"),
	}, 0, 1))

	// чверть століття
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("довгих", "adj:p:v_rod"), atr("чверть", "noun"), atr("століття", "noun:inanim:n:v_rod"),
	}, 0, 1))

	// переконана + m profession
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("переконана", "adj:f:v_naz"), atr("лікар", "noun:anim:m:v_naz"),
	}, 0, 1))

	// станом на
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("чинних", "adj:p:v_rod"), atr("станом", "noun"), atr("на", "prep"),
	}, 0, 1))

	// Богом after pron
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("таку", "adj:f:v_zna:pron"), atr("Богом", "noun:anim:m:v_oru"),
	}, 0, 1))

	// той родом
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atrLemma("той", strPtr("той"), "adj:m:v_naz"), atr("родом", "noun"),
	}, 0, 1))

	// таких
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("таких", "adj:p:v_rod"), atr("люди", "noun:anim:p:v_naz"),
	}, 0, 1))

	// на рівних — Java adjPos > 1 (SENT_START pad)
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT_START"), atr("на", "prep"), atr("рівних", "adj:p:v_rod"), atr("міністри", "noun"),
	}, 2, 3))

	// зразка
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("польські", "adj:p:v_naz"), atr("зразка", "noun"), atr("1620", "number"),
	}, 0, 1))

	// плюс
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("зелених", "adj:p:v_rod"), atr("плюс", "noun"),
	}, 0, 1))

	// пару років
	para := "пара"
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("важкими", "adj:p:v_oru"), atrLemma("пару", &para, "noun:inanim:f:v_zna"),
		atr("років", "noun:inanim:p:v_rod"),
	}, 0, 1))

	// років 6
	rik := "рік"
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("минулих", "adj:p:v_rod"), atrLemma("років", &rik, "noun:inanim:p:v_rod"),
		atr("6", "number"),
	}, 0, 1))

	// осіб на 30
	osoba := "особа"
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("деяких", "adj:p:v_rod"), atrLemma("осіб", &osoba, "noun:anim:p:v_rod"),
		atrLemma("на", strPtr("на"), "prep"), atr("30", "number"),
	}, 0, 1))

	// dash range
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("минулих", "adj:p:v_rod"), atr("травня", "noun:inanim:m:v_rod"),
		atr("–"), atr("липня", "noun:inanim:m:v_rod"),
	}, 0, 1))

	// з + v_oru after plural adj
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("зв'язаних", "adj:p:v_rod"), atr("ченця", "noun:anim:m:v_rod"),
		atr("з", "prep"), atr("черницею", "noun:anim:f:v_oru"),
	}, 0, 1))
}

func TestAdjNounException_ConjAndMore(t *testing.T) {
	// навчальної та середньої шкіл
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT_START"),
		atr("навчальної", "adj:f:v_rod"), atr("та", "conj:coord"),
		atr("середньої", "adj:f:v_rod"), atr("шкіл", "noun:inanim:p:v_rod"),
	}, 3, 4))

	// моїх маму й сестер — need ignore-gender case overlap (v_zna both)
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("моїх", "adj:p:v_zna"), atr("маму", "noun:anim:f:v_zna"),
		atr("й", "conj:coord"), atr("сестер", "noun:anim:p:v_rod"),
	}, 0, 1))

	// коринфський з іонійським ордери — Java adjPos > 2 (SENT pad)
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT_START"),
		atr("коринфський", "adj:m:v_naz"), atr("з", "prep"),
		atr("іонійським", "adj:m:v_oru"), atr("ордери", "noun:inanim:p:v_naz"),
	}, 3, 4))

	// рік тому
	rik := "рік"
	tomu := "тому"
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("пофарбований", "adj:m:v_naz"), atrLemma("рік", &rik, "noun:inanim:m:v_naz"),
		atrLemma("тому", &tomu, "adv"),
	}, 0, 1))

	// два нових горнятка — Java adjPos > 1
	dva := "два"
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT_START"),
		atrLemma("два", &dva, "numr:p:v_naz"), atr("нових", "adj:p:v_rod"),
		atr("горнятка", "noun:inanim:p:v_naz"),
	}, 2, 3))

	// кількох десятих
	des := "десятий"
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("кількох", "numr"), atrLemma("десятих", &des, "adj:p:v_rod"),
		atr("відсотка", "noun:inanim:m:v_rod"),
	}, 1, 2))

	// 2003-го
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("2003-го", "adj:m:v_rod:numr"), atr("прем'єром", "noun"),
	}, 0, 1))

	// 11-ту ранку
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("11-ту", "adj:f:v_zna:numr"), atr("ранку", "noun"),
	}, 0, 1))

	// reverseConjFind2: 3, 4 і 5-ї — left of і is number, right adj numr
	require.True(t, reverseConjFind2([]*languagetool.AnalyzedTokenReadings{
		atr("SENT_START"), atr("3", "number"), atr(","), atr("4", "number"),
		atr("і", "conj"), atr("5-ї", "adj:f:v_rod:numr"), atr("категорій", "noun:inanim:p:v_rod"),
	}, 4, 3))

	// IsNonPluralA
	require.True(t, IsNonPluralA([]*languagetool.AnalyzedTokenReadings{
		atr("а"), atr("просто"),
	}, 0))
	require.False(t, IsNonPluralA([]*languagetool.AnalyzedTokenReadings{
		atr("а"), atrLemma("також", strPtr("також"), "adv"),
	}, 0))
}

func TestAdjNounException_VerbRevAndColors(t *testing.T) {
	// дев'яте травня
	tr := "травень"
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("дев'яте", "adj:n:v_naz:numr"), atrLemma("травня", &tr, "noun:inanim:m:v_rod"),
	}, 0, 1))

	// adjp:actv:bad
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("обмежуючий", "adj:m:v_naz:adjp:actv:bad"), atr("власність", "noun:inanim:f:v_zna"),
	}, 0, 1))

	// нічого + adj inflection overlap
	nishcho := "ніщо"
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT_START"), atr("SENT"), atrLemma("нічого", &nishcho, "noun:inanim:n:v_rod"),
		atr("поганого", "adj:n:v_rod"), atr("людям", "noun:anim:p:v_dav"),
	}, 3, 4))

	// визнання неконституційним закону — Java adjPos > 1, revSearch at adjPos-1
	viz := "визнання"
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT_START"),
		atrLemma("визнання", &viz, "noun:inanim:n:v_naz"),
		atr("неконституційним", "adj:m:v_oru"),
		atr("закону", "noun:inanim:m:v_rod"),
	}, 2, 3))

	// був змушений — revSearch needs startPos > 0 (SENT pad)
	buti := "бути"
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT_START"),
		atrLemma("був", &buti, "verb:imperf:past:m"),
		atr("змушений", "adj:m:v_naz:adjp:pasv"),
		atr("командир", "noun:anim:m:v_naz"),
	}, 2, 3))

	// помальована в біле кімната — adjPos > 2
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT_START"),
		atr("помальована", "adj:f:v_naz:adjp:pasv"), atr("в", "prep"),
		atr("біле", "adj:n:v_zna"), atr("кімната", "noun:inanim:f:v_naz"),
	}, 3, 4))

	// тисячу разів
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("повторена", "adj:f:v_naz:adjp:pasv"), atr("тисячу", "noun"), atr("разів", "noun"),
	}, 0, 1))

	// ще раз
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("покликана", "adj"), atr("ще"), atr("раз", "noun"),
	}, 0, 2))

	// порівняно з попереднім
	poriv := "порівняно"
	z := "з"
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT_START"),
		atrLemma("порівняно", &poriv, "adv"), atrLemma("з", &z, "prep"),
		atr("попереднім", "adj:n:v_oru"), atr("рішенням", "noun"),
	}, 3, 4))

	// GenderMatches
	require.True(t, GenderMatches(
		[]Inflection{{Gender: "m", Case: "v_oru"}},
		[]Inflection{{Gender: "m", Case: "v_rod"}},
		"v_oru", "v_rod"))
}

func TestAdjNounException_FinalArms(t *testing.T) {
	// adjp:pasv + adj:v_oru
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT_START"),
		atr("підсвічений", "adj:m:v_naz:adjp:pasv"),
		atr("синім", "adj:n:v_oru"),
		atr("діамант", "noun:inanim:m:v_naz"),
	}, 2, 3))

	// adjp:pasv + noun v_oru
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("захищені", "adj:p:v_naz:adjp:pasv"), atr("законом", "noun:inanim:m:v_oru"),
	}, 0, 1))

	// adj v_oru + noun v_naz + forward verb
	require.True(t, IsAdjNounException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT_START"), atr("і"),
		atr("Найнижчою", "adj:f:v_oru"), atr("частка", "noun:inanim:f:v_naz"),
		atr("є", "verb:imperf:pres:s:3"),
	}, 2, 3))

	// case government: вдячний + noun + next noun v_rod
	// need adj lemma in case_government map - "вдячний" typically governs v_dav
	// soft: if lemma not in map, caseGovernmentMatches false — use synthetic via HasCaseGovernment
	// Skip if lemma not in map; test hasCaseGovPosRE / caseGovernmentMatches unit-style
	require.True(t, caseGovernmentMatches(
		atrLemma("вдячний", strPtr("вдячний"), "adj:m:v_naz"),
		[]Inflection{{Gender: "m", Case: "v_dav"}}),
		"вдячний should govern v_dav if in case_government map")

	// prev adj governs
	// only if first adj has government of second adj's cases — optional soft skip if map missing

	// TokenSearch verb forward
	require.Equal(t, 2, TokenSearch([]*languagetool.AnalyzedTokenReadings{
		atr("a"), atr("b"), atr("c", "verb:imperf:pres:s:3"),
	}, 1, "verb", nil, nil, DirForward))
}

func TestVerbNounException_MoreJavaArms(t *testing.T) {
	// плюс/мінус
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("додав", "verb"), atr("плюс", "noun"),
	}, 0, 1))

	// 18-го surface
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("закінчилося", "verb"), atr("18-го", "adj"),
	}, 0, 1))

	// всю дорогу
	dor := "дорога"
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("сміялася", "verb"), atr("всю", "adj:f:v_zna"),
		atrLemma("дорогу", &dor, "noun:inanim:f:v_zna"),
	}, 0, 1))

	// impers + v_oru
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("запропоновано", "verb:impers"), atr("відділом", "noun:inanim:m:v_oru"),
	}, 0, 1))

	// кожний v_naz
	kozh := "кожний"
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("займаючись", "verb"), atrLemma("кожен", &kozh, "adj:m:v_naz"),
	}, 0, 1))

	// звалося Proper
	zv := "зватися"
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atrLemma("звалося", &zv, "verb"), atr("Подєбради", "noun:prop"),
	}, 0, 1))

	// тривав + v_zna
	tryv := "тривати"
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atrLemma("тривав", &tryv, "verb"), atr("довгих", "adj:p:v_zna"),
	}, 0, 1))

	// ні впало
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT"), atr("а"), atr("ні"), atr("сіло"), atr("ні"), atr("впало", "verb"),
		atr("щось", "noun"),
	}, 5, 6))

	// не сказати + v_naz
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT"), atr("якщо"), atr("не"), atr("сказати", "verb:inf"),
		atr("слабка", "adj:f:v_naz"),
	}, 3, 4))

	// сортів 10
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("виростили", "verb"), atr("сортів", "noun:inanim:p:v_rod"), atr("10", "number"),
	}, 0, 1))

	// інвестицій на 20
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("залучити", "verb"), atr("інвестицій", "noun:inanim:p:v_rod"),
		atr("на", "prep"), atr("20", "number"),
	}, 0, 1))

	// як боротися підприємцям — Java verbPos > 1
	yak := "як"
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT_START"),
		atrLemma("як", &yak, "adv"), atr("боротися", "verb:imperf:inf"),
		atr("підприємцям", "noun:anim:p:v_dav"),
	}, 2, 3))

	// сміятися гріх
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("сміятися", "verb:imperf:inf"), atr("гріх", "noun"),
	}, 0, 1))

	// брату в обличчя
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("розсміявся", "verb"), atr("брату", "noun:anim:m:v_dav"),
		atr("в", "prep"), atr("обличчя", "noun"),
	}, 0, 1))
}

func TestVerbNounException_MidBlock(t *testing.T) {
	// дай Боже
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("дай", "verb:impr:s:2"), atr("Боже", "noun:anim:m:v_kly"),
	}, 0, 1))

	// fem verb + masc profession
	lem := "лікар"
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("повторила", "verb:perf:past:f"), atrLemma("лікар", &lem, "noun:anim:m:v_naz"),
	}, 0, 1))

	// не існувало + v_rod
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("існувало", "verb:imperf:past:n"), atr("конкуренції", "noun:inanim:f:v_rod"),
	}, 0, 1))

	// меншає людей
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atrLemma("меншає", strPtr("меншати"), "verb:imperf:pres:s:3"),
		atr("людей", "noun:anim:p:v_rod"),
	}, 0, 1))

	// газу менше
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("споживає", "verb"), atr("газу", "noun:inanim:m:v_rod"), atr("менше", "adv"),
	}, 0, 1))

	// небагато надходить книжок — driver before verb
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT_START"), atr("небагато"), atr("надходить", "verb:imperf:pres:s:3"),
		atr("книжок", "noun:inanim:p:v_rod"),
	}, 2, 3))
}

func TestVerbNounException_InfChains(t *testing.T) {
	// V:INF + N + не + V (v_inf gov) — робити прогнозів не вмію
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("робити", "verb:imperf:inf"), atr("прогнозів", "noun:inanim:p:v_rod"),
		atr("не", "part"), atrLemma("вмію", strPtr("вміти"), "verb:imperf:pres:s:1"),
	}, 0, 1))

	// посміхаючись — Java verbPos > 1 (SENT pad); verbPos is the advp
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT_START"),
		atr("пригадує", "verb:imperf:pres:s:3"), atr("посміхаючись", "advp"),
		atr("Аскольд", "noun:anim:m:v_naz:prop"),
	}, 2, 3))

	// працювати неспроможні
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("працювати", "verb:imperf:inf"),
		atrLemma("неспроможні", strPtr("неспроможний"), "adj:p:v_naz"),
	}, 0, 1))

	// ADJ + бути + N v_rod
	require.True(t, IsVerbNounException([]*languagetool.AnalyzedTokenReadings{
		atr("SENT_START"),
		atrLemma("вартий", strPtr("вартий"), "adj:m:v_naz"),
		atrLemma("бути", strPtr("бути"), "verb:imperf:inf"),
		atr("уваги", "noun:inanim:f:v_rod"),
	}, 2, 3))
}
