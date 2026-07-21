package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func frAtr(token, pos, lemma string) *languagetool.AnalyzedTokenReadings {
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

func frSent(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
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

func TestFrenchRepeatedWordsRule_Rule(t *testing.T) {
	rule := NewFrenchRepeatedWordsRule(nil)
	// synonyms: maintenant/A=… — lemma + POS tag matching A
	ss := languagetool.SentenceStartTagName
	s1 := frSent(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		frAtr("Je", "R", "je"),
		frAtr("le", "R", "le"),
		frAtr("fais", "V", "faire"),
		frAtr("maintenant", "A", "maintenant"),
		frAtr(".", ".", "."),
	)
	s2 := frSent(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		frAtr("Et", "C", "et"),
		frAtr("maintenant", "A", "maintenant"),
		frAtr("j'attends", "V", "attendre"),
		frAtr(".", ".", "."),
	)
	matches := rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})
	require.Equal(t, 1, len(matches))
}

func TestFrenchRepeatedWords_IsExceptionAndAdjustPostag(t *testing.T) {
	z := frAtr("Paris", "Z e sp", "Paris")
	require.True(t, frenchRepeatedWordsIsException([]*languagetool.AnalyzedTokenReadings{z}, 0, true, true, false))
	require.Equal(t, "J [me] sp?", frenchRepeatedWordsAdjustPostag("J m s"))
	require.Equal(t, "J [fe] s?p", frenchRepeatedWordsAdjustPostag("J f p"))
}

func TestFrenchRepeatedWords_Messages(t *testing.T) {
	r := NewFrenchRepeatedWordsRule(nil)
	require.Equal(t, "Synonymes de mots répétés.", r.GetDescription())
	require.Equal(t, "Style : Mot répété", r.ShortMsg)
}
