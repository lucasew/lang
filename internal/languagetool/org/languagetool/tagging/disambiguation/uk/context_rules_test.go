package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDisambiguateSt(t *testing.T) {
	start := atrSent("SENT_START", "SENT_START")
	st := atrMulti("ст.", [][2]string{
		{"ст.", "verb:imperf:inf"}, // noise
		{"ст.", "noun:inanim:f:v_naz:nv:abbr:xp1"},
	})
	num := atrSent("208", "number")
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, st, num})
	DisambiguateSt(sent)
	tok := sent.GetTokensWithoutWhitespace()[1]
	require.False(t, tok.HasPartialPosTag("verb"))
	require.True(t, tok.HasPartialPosTag("noun") || tok.HasPartialPosTag("abbr"))
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
	name := atrSent("Бакуліна", "noun:anim:f:v_naz:prop:lname")
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, init, name})
	RetagInitials(sent)
	require.True(t, sent.GetTokensWithoutWhitespace()[1].HasPartialPosTag("fname"))
	require.True(t, sent.GetTokensWithoutWhitespace()[1].HasPartialPosTag("abbr"))
}

func TestHybridAppliesContextRules(t *testing.T) {
	start := atrSent("SENT_START", "SENT_START")
	init := atrUntagged("Є.")
	name := atrSent("Бакуліна", "noun:anim:m:v_rod:prop:lname")
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, init, name})
	out := NewUkrainianHybridDisambiguator().Disambiguate(sent)
	require.True(t, out.GetTokensWithoutWhitespace()[1].HasPartialPosTag("fname"))
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
