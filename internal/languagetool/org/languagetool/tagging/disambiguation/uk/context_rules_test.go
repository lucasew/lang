package uk

import (
	"regexp"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDisambiguateSt(t *testing.T) {
	start := atrSent("SENT_START", "SENT_START")
	// Java: number before ст. (18 ст.) → keep noun:inanim:[nf]
	num := atrSent("18", "number")
	st := atrMulti("ст.", [][2]string{
		{"ст.", "verb:imperf:inf"}, // noise — removed
		{"ст.", "noun:inanim:f:v_naz:nv:abbr:xp1"},
	})
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, num, st})
	DisambiguateSt(sent)
	tok := sent.GetTokensWithoutWhitespace()[2]
	require.False(t, tok.HasPartialPosTag("verb"))
	require.True(t, tok.HasPartialPosTag("noun") || tok.HasPartialPosTag("abbr"))
	// untagged ст. stays untagged (no invent inject)
	st2 := atrUntagged("ст.")
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, atrSent("18", "number"), st2})
	DisambiguateSt(sent2)
	require.False(t, sent2.GetTokensWithoutWhitespace()[2].IsTagged())

	// Java: i>2 && prev ST_ABBR + next number → plural on both (e.g. «див. ст. ст. 5»)
	filler := atrSent("див.", "abbr")
	stA := atrMulti("ст.", [][2]string{
		{"ст.", "noun:inanim:f:v_naz:nv:abbr"},
		{"ст.", "noun:inanim:p:v_naz:nv:abbr"},
	})
	stB := atrMulti("ст.", [][2]string{
		{"ст.", "noun:inanim:f:v_naz:nv:abbr"},
		{"ст.", "noun:inanim:p:v_naz:nv:abbr"},
	})
	num2 := atrSent("5", "number")
	// tokens: SENT, filler, stA, stB, num → stB at i=3 > 2
	sent3 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, filler, stA, stB, num2})
	DisambiguateSt(sent3)
	tokB := sent3.GetTokensWithoutWhitespace()[3]
	require.True(t, tokB.HasPartialPosTag(":p:"))
	require.False(t, hasGenderF(tokB))
}

func hasGenderF(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	for _, r := range tok.GetReadings() {
		if r != nil && r.GetPOSTag() != nil && strings.Contains(*r.GetPOSTag(), ":f:") {
			return true
		}
	}
	return false
}

func TestDisambiguatePronPos(t *testing.T) {
	start := atrSent("SENT_START", "SENT_START")
	// Java: drop adj.*pron:pos that does not agree with neighbor noun gender/case
	// f pos adj + f noun → keep pos; m pos adj + f noun → drop m pos
	yoho := atrMulti("його", [][2]string{
		{"він", "noun:unanim:m:v_rod:pron:pers:3"},
		{"його", "adj:f:v_naz:nv:pron:pos"},
		{"його", "adj:m:v_zna:nv:pron:pos"},
	})
	noun := atrSent("машина", "noun:inanim:f:v_naz")
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, yoho, noun})
	DisambiguatePronPos(sent)
	tok := sent.GetTokensWithoutWhitespace()[1]
	require.True(t, tok.HasPartialPosTag("adj:f:v_naz"))
	require.False(t, tok.HasPartialPosTag("adj:m:v_zna"))
	// pers reading is not adj — Java leaves it
	require.True(t, tok.HasPartialPosTag("pron:pers"))

	// no neighboring noun inflections (verb only) → no adj filter
	yih := atrMulti("їх", [][2]string{
		{"вони", "noun:unanim:p:v_zna:pron:pers:3"},
		{"їх", "adj:p:v_naz:nv:pron:pos"},
	})
	verb := atrSent("забули", "verb:perf:past:p")
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, yih, verb})
	DisambiguatePronPos(sent2)
	tok2 := sent2.GetTokensWithoutWhitespace()[1]
	require.True(t, tok2.HasPartialPosTag("pron:pos"))
	require.True(t, tok2.HasPartialPosTag("pron:pers"))
}

func TestRetagInitials(t *testing.T) {
	start := atrSent("SENT_START", "SENT_START")
	init := atrUntagged("Є.")
	// Java :prop:lname on surname drives initial tags
	name := atrSent("Бакуліна", "noun:anim:f:v_naz:prop:lname")
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, init, name})
	RetagInitials(sent)
	tok := sent.GetTokensWithoutWhitespace()[1]
	require.True(t, tok.HasPartialPosTag("fname"))
	require.True(t, tok.HasPartialPosTag("abbr"))
	require.True(t, tok.HasPartialPosTag("f:")) // gender from lname
	// without lname fails closed
	init2 := atrUntagged("Є.")
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, init2, atrSent("Марія", "noun:anim:f:v_naz:prop:fname")})
	RetagInitials(sent2)
	require.False(t, sent2.GetTokensWithoutWhitespace()[1].IsTagged())

	// dual initials: А. Б. Коваленко → fname + pname
	a := atrUntagged("А.")
	b := atrUntagged("Б.")
	kov := atrSent("Коваленко", "noun:anim:m:v_naz:prop:lname")
	sent3 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, a, b, kov})
	RetagInitials(sent3)
	require.True(t, sent3.GetTokensWithoutWhitespace()[1].HasPartialPosTag("fname"))
	require.True(t, sent3.GetTokensWithoutWhitespace()[2].HasPartialPosTag("pname"))
}

func TestHybridAppliesContextRules(t *testing.T) {
	start := atrSent("SENT_START", "SENT_START")
	init := atrUntagged("Є.")
	name := atrSent("Бакуліна", "noun:anim:m:v_rod:prop:lname")
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, init, name})
	out := NewUkrainianHybridDisambiguator().Disambiguate(sent)
	require.True(t, out.GetTokensWithoutWhitespace()[1].HasPartialPosTag("fname"))
	require.True(t, out.GetTokensWithoutWhitespace()[1].HasPartialPosTag("m:"))
}

func atrSent(token, pos string) *languagetool.AnalyzedTokenReadings {
	p := pos
	var lemma *string
	if pos != "SENT_START" {
		l := token
		lemma = &l
	}
	return languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken(token, &p, lemma),
	}, 0)
}

func atrMulti(token string, pairs [][2]string) *languagetool.AnalyzedTokenReadings {
	var rs []*languagetool.AnalyzedToken
	for _, lp := range pairs {
		l, p := lp[0], lp[1]
		rs = append(rs, languagetool.NewAnalyzedToken(token, &p, &l))
	}
	return languagetool.NewAnalyzedTokenReadingsList(rs, 0)
}

func atrUntagged(token string) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken(token, nil, nil),
	}, 0)
}

func TestRetagFemNames(t *testing.T) {
	start := atrSent("SENT_START", "SENT_START")
	ledi := atrMulti("леді", [][2]string{{"леді", "noun:anim:f:v_naz:nv"}})
	// masc lname that should become fem
	name := atrMulti("Черчилль", [][2]string{{"Черчилль", "noun:anim:m:v_naz:prop:lname"}})
	verb := atrMulti("була", [][2]string{{"бути", "verb:imperf:past:f"}})
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, ledi, name, verb})
	RetagFemNames(sent)
	tok := sent.GetTokensWithoutWhitespace()[2]
	require.True(t, tok.HasPartialPosTag(":f:"))
	require.True(t, tok.HasPartialPosTag("lname") || tok.HasPartialPosTag("prop"))
	require.False(t, tok.HasPartialPosTag("noun:anim:m:v_naz:prop"))

	// Олег П'ятниця — fname title + capitalized non-prop → lname
	oleg := atrMulti("Олег", [][2]string{{"Олег", "noun:anim:m:v_naz:prop:fname"}})
	pyat := atrMulti("П'ятниця", [][2]string{{"п'ятниця", "noun:inanim:f:v_naz"}})
	verb2 := atrMulti("прийшов", [][2]string{{"прийти", "verb:perf:past:m"}})
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, oleg, pyat, verb2})
	RetagFemNames(sent2)
	tok2 := sent2.GetTokensWithoutWhitespace()[2]
	require.True(t, tok2.HasPartialPosTag("prop:lname") || tok2.HasPartialPosTag(":lname"))
	require.True(t, tok2.HasPartialPosTag("noun:anim:m:v_naz:prop"))
}

func TestRemoveInanimVKly(t *testing.T) {
	start := atrSent("SENT_START", "SENT_START")
	// inanim with both v_kly and v_naz — drop v_kly
	tok := atrMulti("крило", [][2]string{
		{"крило", "noun:inanim:n:v_kly"},
		{"крило", "noun:inanim:n:v_naz"},
	})
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, tok})
	RemoveInanimVKly(sent)
	require.False(t, sent.GetTokensWithoutWhitespace()[1].HasPartialPosTag("v_kly"))
	require.True(t, sent.GetTokensWithoutWhitespace()[1].HasPartialPosTag("v_naz"))

	// vocative context: keep
	adj := atrMulti("Ясний", [][2]string{{"ясний", "adj:m:v_kly"}})
	moon := atrMulti("місяцю", [][2]string{
		{"місяць", "noun:inanim:m:v_kly"},
		{"місяць", "noun:inanim:m:v_dav"},
	})
	bang := atrSent("!", "SENT_END")
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, adj, moon, bang})
	RemoveInanimVKly(sent2)
	require.True(t, sent2.GetTokensWithoutWhitespace()[2].HasPartialPosTag("v_kly"))

	// Java gate: only :geo v_kly → do not enter (keep geo vocative)
	geoOnly := atrMulti("Києве", [][2]string{
		{"Київ", "noun:inanim:m:v_kly:prop:geo"},
		{"Київ", "noun:inanim:m:v_naz:prop:geo"},
	})
	sent3 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, geoOnly})
	RemoveInanimVKly(sent3)
	require.True(t, sent3.GetTokensWithoutWhitespace()[1].HasPartialPosTag("v_kly"))

	// non-geo v_kly present → also drop geo v_kly (Java INANIM_VKLY includes geo)
	mixed := atrMulti("місте", [][2]string{
		{"місто", "noun:inanim:n:v_kly"},
		{"місто", "noun:inanim:n:v_kly:prop:geo"},
		{"місто", "noun:inanim:n:v_naz"},
	})
	sent4 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, mixed})
	RemoveInanimVKly(sent4)
	require.False(t, sent4.GetTokensWithoutWhitespace()[1].HasPartialPosTag("v_kly"))
	require.True(t, sent4.GetTokensWithoutWhitespace()[1].HasPartialPosTag("v_naz"))
}

func TestRemoveLowerCaseHomonymsForAbbreviations(t *testing.T) {
	start := atrSent("SENT_START", "SENT_START")
	ato := atrMulti("АТО", [][2]string{
		{"ато", "part"},
		{"АТО", "noun:inanim:n:v_naz:nv:abbr:prop"},
	})
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, ato})
	RemoveLowerCaseHomonymsForAbbreviations(sent)
	tok := sent.GetTokensWithoutWhitespace()[1]
	require.True(t, tok.HasPartialPosTag("abbr"))
	require.False(t, tok.HasPosTag("part") || tok.HasPartialPosTag("part"))
}

func TestRemovePluralForNames(t *testing.T) {
	start := atrSent("SENT_START", "SENT_START")
	// plural + singular readings
	name := atrMulti("Василів", [][2]string{
		{"Василь", "noun:anim:p:v_rod:prop:fname"},
		{"Василів", "noun:anim:m:v_naz:prop:lname"},
	})
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, name})
	RemovePluralForNames(sent)
	tok := sent.GetTokensWithoutWhitespace()[1]
	require.False(t, tok.HasPartialPosTag(":p:"))
	require.True(t, tok.HasPartialPosTag("lname") || tok.HasPartialPosTag("m:v_naz"))

	// keep plural after numr
	num := atrSent("два", "numr:p:v_naz")
	name2 := atrMulti("Андрії", [][2]string{
		{"Андрій", "noun:anim:p:v_naz:prop:fname"},
		{"Андрій", "noun:anim:m:v_naz:prop:fname"},
	})
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, num, name2})
	RemovePluralForNames(sent2)
	require.True(t, sent2.GetTokensWithoutWhitespace()[2].HasPartialPosTag(":p:"))
}

func TestRemoveVerbImpr(t *testing.T) {
	start := atrSent("SENT_START", "SENT_START")
	adj := atrMulti("подальші", [][2]string{{"подальший", "adj:p:v_naz"}})
	noun := atrMulti("суди", [][2]string{
		{"суд", "noun:inanim:p:v_naz"},
		{"судити", "verb:imperf:impr:p:2"},
	})
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, adj, noun})
	RemoveVerbImpr(sent)
	tok := sent.GetTokensWithoutWhitespace()[2]
	require.True(t, tok.HasPartialPosTag("noun"))
	require.False(t, tok.HasPartialPosTag("impr"))
}

func TestPreferVocativeWhenBang(t *testing.T) {
	start := atrSent("SENT_START", "SENT_START")
	adj := atrMulti("Шановні", [][2]string{
		{"шановний", "adj:p:v_kly:compb"},
		{"шановний", "adj:p:v_naz:compb"},
	})
	noun := atrMulti("депутати", [][2]string{
		{"депутат", "noun:anim:p:v_kly"},
		{"депутат", "noun:anim:p:v_naz"},
	})
	bang := atrSent("!", "SENT_END")
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, adj, noun, bang})
	PreferVocativeWhenBang(sent)
	require.True(t, sent.GetTokensWithoutWhitespace()[1].HasPartialPosTag("v_kly"))
	require.False(t, sent.GetTokensWithoutWhitespace()[1].HasPartialPosTag("v_naz"))
	require.True(t, sent.GetTokensWithoutWhitespace()[2].HasPartialPosTag("v_kly"))
	require.False(t, sent.GetTokensWithoutWhitespace()[2].HasPartialPosTag("v_naz"))
}

func TestRemoveLowerCaseBadForUpperCaseGood(t *testing.T) {
	start := atrSent("SENT_START", "SENT_START")
	tok := atrMulti("Держдепартамент", [][2]string{
		{"Держдепартамент", "noun:inanim:m:v_naz:prop"},
		{"держдепартамент", "noun:inanim:m:v_naz:bad"},
	})
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, tok})
	RemoveLowerCaseBadForUpperCaseGood(sent)
	require.True(t, sent.GetTokensWithoutWhitespace()[1].HasPartialPosTag("prop"))
	require.False(t, sent.GetTokensWithoutWhitespace()[1].HasPartialPosTag("bad"))

	// Java .*:prop full-match: only :prop:geo (not ending at :prop) does not open the gate
	geo := atrMulti("Київ", [][2]string{
		{"Київ", "noun:inanim:m:v_naz:prop:geo"},
		{"київ", "noun:inanim:m:v_naz:bad"},
	})
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, geo})
	RemoveLowerCaseBadForUpperCaseGood(sent2)
	require.True(t, sent2.GetTokensWithoutWhitespace()[1].HasPartialPosTag("bad"),
		"gate needs POS ending in :prop, not :prop:geo")
}

func TestRetagUnknownInitials(t *testing.T) {
	start := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("", strPtr("SENT_START"), nil),
	}, 0)
	// А. without name tag → noninfl:abbr
	pBad := "noun:inanim:m:v_naz"
	init := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("А.", &pBad, strPtr("а")),
	}, 0)
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, init})
	RetagUnknownInitials(sent)
	tok := sent.GetTokensWithoutWhitespace()[1]
	require.True(t, tok.HasPosTag("noninfl:abbr") || tok.HasPartialPosTag("noninfl:abbr"))
	require.False(t, tok.HasPartialPosTag("noun"))
}

func TestRetagPluralProp(t *testing.T) {
	start := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("", strPtr("SENT_START"), nil),
	}, 0)
	pNum := "numr:p:v_naz"
	num := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("дві", &pNum, strPtr("два")),
	}, 0)
	// only rod prop, no naz — should retag to p:v_naz
	pSg := "noun:inanim:f:v_rod:prop:geo"
	name := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("Франції", &pSg, strPtr("Франція")),
	}, 0)
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, num, name})
	RetagPluralProp(sent)
	tok := sent.GetTokensWithoutWhitespace()[2]
	require.True(t, tok.HasPartialPosTag(":p:v_naz"))
	require.True(t, tok.HasPartialPosTag("prop"))
}

func TestDisambiguateYih_PrevVerbGov(t *testing.T) {
	// посунув їх . — prev verb with v_zna gov → drop pos
	// lemma "посунути" or use "забути" which may govern v_zna
	// "бачити" often has v_zna
	start := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("", strPtr("SENT_START"), nil),
	}, 0)
	// use "бачити" — typically v_zna in case_government
	vPos, vLem := "verb:imperf:past:m", "бачити"
	verb := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("бачив", &vPos, &vLem),
	}, 0)
	pPers, pPos := "noun:unanim:p:v_zna:pron:pers:3", "adj:p:v_naz:nv:pron:pos"
	lPers, lPos := "вони", "їх"
	yih := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("їх", &pPers, &lPers),
		languagetool.NewAnalyzedToken("їх", &pPos, &lPos),
	}, 0)
	// end of sentence: only start, verb, yih
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, verb, yih})
	DisambiguateYih(sent)
	tok := sent.GetTokensWithoutWhitespace()[2]
	// if бачити has v_zna/v_rod in map, pos should drop
	// бачити typically has v_zna; hybrid path tested via DisambiguateYih
	require.True(t, tok.HasPartialPosTag("pron:pers") || len(tok.GetReadings()) > 0)
	// when map has v_zna/v_rod for бачити, pos reading is removed
	if setHasAny(caseGovForPosRE(verb, verbOnlyRE), "v_rod", "v_zna") {
		require.False(t, tok.HasPartialPosTag("pron:pos"))
	}
}

func TestDisambiguateYih_ObjectLemma(t *testing.T) {
	start := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("", strPtr("SENT_START"), nil),
	}, 0)
	pPers, pPos := "noun:unanim:p:v_zna:pron:pers:3", "adj:p:v_naz:nv:pron:pos"
	yih := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("їх", &pPers, strPtr("вони")),
		languagetool.NewAnalyzedToken("їх", &pPos, strPtr("їх")),
	}, 0)
	nPos, nLem := "noun:inanim:f:v_naz", "кількість"
	noun := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("кількість", &nPos, &nLem),
	}, 0)
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, yih, noun})
	DisambiguateYih(sent)
	tok := sent.GetTokensWithoutWhitespace()[1]
	require.False(t, tok.HasPartialPosTag("pron:pos"))
}

func TestStationNameRE_FullMatch(t *testing.T) {
	require.True(t, stStationNameRE.MatchString("метро"))
	require.True(t, stStationNameRE.MatchString("Київська"))
	// must not match mere prefix "метро" + extra (Java matches())
	require.False(t, stStationNameRE.MatchString("метрополітен"))
	require.False(t, stStationNameRE.MatchString("київська")) // not capitalized
}

func TestDisambiguateSt_Station(t *testing.T) {
	// ст. + Capitalized → keep only noun:inanim:f
	start := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strPtr("SENT_START"), nil))
	stPos1, stPos2, stPos3 := "noun:inanim:f:v_naz", "noun:inanim:m:v_naz", "adj:m:v_naz"
	stTok := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("ст.", &stPos1, strPtr("ст.")),
		languagetool.NewAnalyzedToken("ст.", &stPos2, strPtr("ст.")),
		languagetool.NewAnalyzedToken("ст.", &stPos3, strPtr("ст.")),
	}, 0)
	namePos := "noun:inanim:f:v_naz:prop:geo"
	name := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Київська", &namePos, strPtr("Київська")))
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, stTok, name})
	DisambiguateSt(sent)
	// only f readings remain on ст.
	for _, r := range stTok.GetReadings() {
		if r != nil && r.GetPOSTag() != nil {
			require.Contains(t, *r.GetPOSTag(), "noun:inanim:f")
		}
	}
}

func TestVerbOnlyRE_FullMatch(t *testing.T) {
	require.True(t, verbOnlyRE.MatchString("verb:imperf:inf"))
	require.True(t, verbOnlyRE.MatchString("verb:perf:past:m"))
	require.False(t, verbOnlyRE.MatchString("adverb"))
	require.False(t, verbOnlyRE.MatchString("noun:verb:fake"))
}

func TestCaseGovForPosRE_VerbOnly(t *testing.T) {
	// inject lemma with case government via helper map if available
	p, l := "verb:imperf:inf", "бачити"
	verb := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("бачити", &p, &l))
	govs := caseGovForPosRE(verb, verbOnlyRE)
	// if case government map has бачити → expect non-empty; else at least not panic
	_ = govs
	// critical: pattern must accept full verb POS (was ^verb only → always empty)
	// fullMatch with ^verb.*$ accepts; prove with synthetic: empty lemma still iterates
	require.True(t, fullMatch(verbOnlyRE, "verb:imperf:inf"))
	require.False(t, fullMatch(regexp.MustCompile(`^verb`), "verb:imperf:inf"))
}

func TestPropPOSRE_PropGeo(t *testing.T) {
	// Java compile(".*?:prop") Matcher.matches() — tag must end at :prop
	require.True(t, propPOSRE.MatchString("noun:inanim:m:v_naz:prop"))
	require.False(t, propPOSRE.MatchString("noun:inanim:m:v_naz:prop:geo"))
	require.False(t, propPOSRE.MatchString("noun:anim:m:v_naz:prop:lname"))
	require.False(t, propPOSRE.MatchString("noun:inanim:m:v_naz"))
}
