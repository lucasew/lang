package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCountLineBreaks(t *testing.T) {
	require.Equal(t, 0, CountLineBreaks("abc"))
	require.Equal(t, 2, CountLineBreaks("a\nb\nc"))
}

func TestProcessColumnChange(t *testing.T) {
	require.Equal(t, 5, ProcessColumnChange(2, "abc")) // 2+3
	require.Equal(t, 2, ProcessColumnChange(1, "ab\nc")) // len-lastIndex(\\n) = 4-2
}

func TestFindLineColumnInSentences(t *testing.T) {
	sents := []SentenceData{
		NewSentenceData(nil, "Hello. ", 0, 0, 0),
		NewSentenceData(nil, "World", 7, 0, 7),
	}
	// offset 8 → second sentence, 'o' of World → prefix "W"
	p := FindLineColumnInSentences(sents, 8)
	require.Equal(t, 0, p.Line)
	require.Equal(t, 8, p.Column) // startColumn 7 + len("W") = 8
}
