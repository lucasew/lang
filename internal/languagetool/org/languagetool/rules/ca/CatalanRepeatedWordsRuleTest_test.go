package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func caAtr(token, pos, lemma string) *languagetool.AnalyzedTokenReadings {
	var p, l *string
	if pos != "" {
		pp := pos
		p = &pp
	}
	if lemma != "" {
		ll := lemma
		l = &ll
	}
	return languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(token, p, l), 0)
}

func caSent(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	pos := 0
	for _, t := range toks {
		if t == nil {
			continue
		}
		t.SetStartPos(pos)
		pos += len([]rune(t.GetToken())) + 1
	}
	return languagetool.NewAnalyzedSentence(toks)
}

func TestCatalanRepeatedWordsRule(t *testing.T) {
	rule := NewCatalanRepeatedWordsRule(nil)
	// Java: real lemmas (synonyms key "suggerir")
	ss := languagetool.SentenceStartTagName
	s1 := caSent(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		caAtr("Vull", "V", "voler"),
		caAtr("suggerir", "V", "suggerir"),
		caAtr("una", "D", "un"),
		caAtr("idea", "N", "idea"),
		caAtr(".", ".", "."),
	)
	s2 := caSent(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		caAtr("Puc", "V", "poder"),
		caAtr("suggerir", "V", "suggerir"),
		caAtr("una", "D", "un"),
		caAtr("altra", "D", "altre"),
		caAtr(".", ".", "."),
	)
	require.Equal(t, 1, len(rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})))
}

func TestCatalanRepeatedWords_IsExceptionAndAdjustPostag(t *testing.T) {
	np := caAtr("Barcelona", "NP00000", "Barcelona")
	require.True(t, catalanRepeatedWordsIsException([]*languagetool.AnalyzedTokenReadings{np}, 0, true, true, false))
	// CA CS → [MFC][SN] (differs from ES)
	require.Equal(t, "NC[MFC][SN]000", catalanRepeatedWordsAdjustPostag("NCCS000"))
}

func TestCatalanRepeatedWords_Messages(t *testing.T) {
	r := NewCatalanRepeatedWordsRule(nil)
	require.Equal(t, "Sinònims per a paraules repetides.", r.GetDescription())
	require.Equal(t, "Estil: paraula repetida", r.ShortMsg)
}

func TestCatalanRepeatedWords_AntiPatternsCount(t *testing.T) {
	require.Equal(t, 1, len(CatalanRepeatedWordsAntiPatterns), "Java ANTI_PATTERNS 1/1")
}
