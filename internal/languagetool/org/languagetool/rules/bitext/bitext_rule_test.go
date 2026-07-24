package bitext

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func sentence(text string) *languagetool.AnalyzedSentence {
	return languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(text, nil, nil), 0),
	})
}

func multiWordSentence(words ...string) *languagetool.AnalyzedSentence {
	var toks []*languagetool.AnalyzedTokenReadings
	pos := 0
	for _, w := range words {
		toks = append(toks, languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(w, nil, nil), pos))
		pos += len(w) + 1
	}
	return languagetool.NewAnalyzedSentence(toks)
}

func TestDifferentLengthRule(t *testing.T) {
	r := NewDifferentLengthRule()
	src := sentence("hello")
	trg := sentence("x")
	m := r.MatchBitext(src, trg)
	require.NotEmpty(t, m)
	require.Equal(t, "TRANSLATION_LENGTH", r.GetID())
}

func TestSameTranslationRule(t *testing.T) {
	r := NewSameTranslationRule()
	// need >3 nws tokens and same text
	words := []string{"one", "two", "three", "four"}
	src := multiWordSentence(words...)
	// same tokens → same GetText join
	trg := multiWordSentence(words...)
	require.Equal(t, src.GetText(), trg.GetText())
	m := r.MatchBitext(src, trg)
	require.NotEmpty(t, m)
}

func TestDifferentPunctuationRule(t *testing.T) {
	r := NewDifferentPunctuationRule()
	src := multiWordSentence("Hi", ".")
	trg := multiWordSentence("Hi", "!")
	m := r.MatchBitext(src, trg)
	require.NotEmpty(t, m)
}

func TestRelevantBitextRules(t *testing.T) {
	require.Len(t, RelevantBitextRules(), 3)
}

func TestSelectBitextRules_Builtin(t *testing.T) {
	all := RelevantBitextRules()
	require.NotEmpty(t, all)
	first := all[0].GetID()
	got := SelectBitextRules(all, []string{first}, nil, false)
	require.Len(t, got, len(all)-1)
	for _, r := range got {
		require.NotEqual(t, first, r.GetID())
	}
	got2 := SelectBitextRules(all, nil, []string{first}, true)
	require.Len(t, got2, 1)
	require.Equal(t, first, got2[0].GetID())
}
