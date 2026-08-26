package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestIsVerbDicendi(t *testing.T) {
	require.True(t, IsVerbDicendi("dir"))
	require.True(t, IsVerbDicendi("explicar"))
	require.True(t, IsVerbDicendi("DIR"))
	require.False(t, IsVerbDicendi("correr"))
	require.Equal(t, 168, len(verbsDicendi))
}

func atrLemmaPos(lemma, pos string) *languagetool.AnalyzedTokenReadings {
	var p, l *string
	if pos != "" {
		pp := pos
		p = &pp
	}
	if lemma != "" {
		ll := lemma
		l = &ll
	}
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("x", p, l), 0)
}

func TestIsVerbDicendiBeforeTokens(t *testing.T) {
	// ... ell va dir que ...
	// index: 0 start, 1 ell, 2 va(V), 3 dir(V lemma), 4 que
	// Looking from index 4: que not keepLooking → false without going back
	// Looking from 3: dir is V + dicendi → true
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrLemmaPos("", "SENT_START"),
		atrLemmaPos("ell", "PP3MS000"),
		atrLemmaPos("anar", "VMI3S0"), // V.*
		atrLemmaPos("dir", "VMN0000"),
		atrLemmaPos("que", "CS"),
	}
	require.True(t, IsVerbDicendiBeforeTokens(tokens, 3))
	// from 4: "que" has no V/RG match → false
	require.False(t, IsVerbDicendiBeforeTokens(tokens, 4))
	// adverb RG then dicendi
	tokens2 := []*languagetool.AnalyzedTokenReadings{
		atrLemmaPos("", "SENT_START"),
		atrLemmaPos("dir", "VMN0000"),
		atrLemmaPos("molt", "RG"),
	}
	// from 2: RG keep looking, then dir dicendi
	require.True(t, IsVerbDicendiBeforeTokens(tokens2, 2))
}
