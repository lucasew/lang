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
	require.Equal(t, 2, ProcessColumnChange(1, "ab\nc")) // UTF-16 len-lastIndex(\\n) = 4-2
	// Non-ASCII after last newline: Java String.length() is UTF-16, not UTF-8 bytes.
	// "a\né" → length 3, lastIndex('\\n')=1 → column 2 (not byte-len 4-1=3).
	require.Equal(t, 2, ProcessColumnChange(0, "a\né"))
	// "aé" no newline → 0 + UTF-16 length 2
	require.Equal(t, 2, ProcessColumnChange(0, "aé"))
	// singleLineBreaksMarksPara=false and sentence starts with \\n → column--
	require.Equal(t, 0, ProcessColumnChangePara(0, "\n", false)) // len1 - 0 - 1 = 0
	require.Equal(t, 1, ProcessColumnChangePara(0, "\n", true))  // len1 - 0 = 1
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

	// UTF-16 offsets into multi-byte text: c a f é \n X → units 0..5
	sents2 := []SentenceData{
		NewSentenceData(nil, "café\nX", 0, 0, 0),
	}
	// offset 4 is '\\n'; prefix "café" → line 0, column 4
	p2 := FindLineColumnInSentences(sents2, 4)
	require.Equal(t, 0, p2.Line)
	require.Equal(t, 4, p2.Column)
	// offset 5 is 'X' after newline → line 1
	p3 := FindLineColumnInSentences(sents2, 5)
	require.Equal(t, 1, p3.Line)
	require.Equal(t, 1, p3.Column) // prefix "café\\nX" lastIndex nl; ProcessColumnChange → 1
}
