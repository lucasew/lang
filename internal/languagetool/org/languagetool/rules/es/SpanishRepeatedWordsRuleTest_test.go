package es

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func esAtr(token, pos, lemma string) *languagetool.AnalyzedTokenReadings {
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

func esSent(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
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

func TestSpanishRepeatedWordsRule_Rule(t *testing.T) {
	rule := NewSpanishRepeatedWordsRule(nil)
	// Java: real lemmas (synonyms key "sugerir")
	ss := languagetool.SentenceStartTagName
	s1 := esSent(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		esAtr("Voy", "V", "ir"),
		esAtr("a", "P", "a"),
		esAtr("sugerir", "V", "sugerir"),
		esAtr("algo", "P", "algo"),
		esAtr(".", ".", "."),
	)
	s2 := esSent(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		esAtr("Puedo", "V", "poder"),
		esAtr("sugerir", "V", "sugerir"),
		esAtr("otra", "D", "otro"),
		esAtr("cosa", "N", "cosa"),
		esAtr(".", ".", "."),
	)
	require.Equal(t, 1, len(rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})))
}

func TestSpanishRepeatedWords_IsExceptionAndAdjustPostag(t *testing.T) {
	np := esAtr("Madrid", "NP00000", "Madrid")
	require.True(t, spanishRepeatedWordsIsException([]*languagetool.AnalyzedTokenReadings{np}, 0, true, true, false))
	require.Equal(t, "NC[MC][SN]000", spanishRepeatedWordsAdjustPostag("NCMS000"))
	require.Equal(t, "NC[FC][SN]000", spanishRepeatedWordsAdjustPostag("NCFS000"))
}

func TestSpanishRepeatedWords_Messages(t *testing.T) {
	r := NewSpanishRepeatedWordsRule(nil)
	require.Equal(t, "Sinónimos para palabras repetidas.", r.GetDescription())
	require.Equal(t, "Estilo: palabra repetida", r.ShortMsg)
	require.Equal(t, 1, r.MinToCheckParagraph())
}

func TestSpanishRepeatedWords_AntiPatternsCount(t *testing.T) {
	require.Equal(t, 5, len(SpanishRepeatedWordsAntiPatterns), "Java ANTI_PATTERNS 5/5")
}
