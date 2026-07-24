package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsParagraphEnd(t *testing.T) {
	a := AnalyzePlain("para one\n\n")
	b := AnalyzePlain("para two")
	sents := []*AnalyzedSentence{a, b}
	require.True(t, IsParagraphEnd(sents, 0, false))
	require.True(t, IsParagraphEnd(sents, 1, false))

	c := AnalyzePlain("line\n")
	d := AnalyzePlain("next")
	require.True(t, IsParagraphEnd([]*AnalyzedSentence{c, d}, 0, true))

	e := AnalyzePlain("x")
	f := AnalyzePlain("\nmore")
	require.True(t, IsParagraphEnd([]*AnalyzedSentence{e, f}, 0, false))
}
