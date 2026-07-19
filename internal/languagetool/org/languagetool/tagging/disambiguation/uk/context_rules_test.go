package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDisambiguateSt(t *testing.T) {
	start := atrSent("SENT_START", "SENT_START")
	// Java: number before ст. (18 ст.)
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
}

func TestDisambiguatePronPos(t *testing.T) {
	start := atrSent("SENT_START", "SENT_START")
	// його + noun → drop pers, keep pos
	yoho := atrMulti("його", [][2]string{
		{"він", "noun:unanim:m:v_rod:pron:pers:3"},
		{"його", "adj:f:v_naz:nv:pron:pos"},
	})
	noun := atrSent("машина", "noun:inanim:f:v_naz")
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, yoho, noun})
	DisambiguatePronPos(sent)
	tok := sent.GetTokensWithoutWhitespace()[1]
	require.True(t, tok.HasPartialPosTag("pron:pos"))
	require.False(t, tok.HasPartialPosTag("pron:pers"))

	// їх + verb → drop pos
	yih := atrMulti("їх", [][2]string{
		{"вони", "noun:unanim:p:v_zna:pron:pers:3"},
		{"їх", "adj:p:v_naz:nv:pron:pos"},
	})
	verb := atrSent("забули", "verb:perf:past:p")
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, yih, verb})
	DisambiguatePronPos(sent2)
	tok2 := sent2.GetTokensWithoutWhitespace()[1]
	require.True(t, tok2.HasPartialPosTag("pron:pers"))
	require.False(t, tok2.HasPartialPosTag("pron:pos"))
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
}
