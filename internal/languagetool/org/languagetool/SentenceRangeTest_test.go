package languagetool

import (
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/markup"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.SentenceRangeTest.

func TestSentenceRange_CorrectSentenceRange(t *testing.T) {
	sentences := []string{
		"Hallo,\n\n",
		"Das ist ein neuer Satz.",
		"\n\nEin Satz mit \uFEFFSonderzeichen.",
		"\n\n\n\n\nSatz mehreren Leerzeichen.",
		" Hier sind die Zeichen mal am Ende.\n\n\n",
		"\n\n\n\uFEFFNoch ein Satz.\n\n\n\n",
	}
	text := strings.Join(sentences, "")
	annotatedText := markup.NewAnnotatedTextBuilder().AddText(text).Build()
	ranges := GetRangesFromSentences(annotatedText, sentences)
	require.Len(t, ranges, 6)

	require.Equal(t, 0, ranges[0].GetFromPos())
	require.Equal(t, 6, ranges[0].GetToPos())
	require.Equal(t, 8, ranges[1].GetFromPos())
	require.Equal(t, 31, ranges[1].GetToPos())
	require.Equal(t, 33, ranges[2].GetFromPos())
	require.Equal(t, 61, ranges[2].GetToPos())
	require.Equal(t, 66, ranges[3].GetFromPos())
	require.Equal(t, 92, ranges[3].GetToPos())
	require.Equal(t, 93, ranges[4].GetFromPos())
	require.Equal(t, 127, ranges[4].GetToPos())
	require.Equal(t, 133, ranges[5].GetFromPos())
	require.Equal(t, 148, ranges[5].GetToPos())

	var sb strings.Builder
	for _, sr := range ranges {
		sb.WriteString(utf16SubstrSR(text, sr.GetFromPos(), sr.GetToPos()))
	}
	require.Equal(t,
		"Hallo,Das ist ein neuer Satz.Ein Satz mit \uFEFFSonderzeichen.Satz mehreren Leerzeichen.Hier sind die Zeichen mal am Ende.\uFEFFNoch ein Satz.",
		sb.String())
}

func utf16SubstrSR(s string, from, to int) string {
	u := utf16.Encode([]rune(s))
	if from < 0 {
		from = 0
	}
	if to > len(u) {
		to = len(u)
	}
	if from >= to {
		return ""
	}
	return string(utf16.Decode(u[from:to]))
}

func TestSentenceRange_GermanSentenceRange(t *testing.T) {
	t.Skip("unimplemented: SentenceRangeTest.testGermanSentenceRange")
}
func TestSentenceRange_EnglishSentenceRange(t *testing.T) {
	t.Skip("unimplemented: SentenceRangeTest.testEnglishSentenceRange")
}
func TestSentenceRange_SpecialCase(t *testing.T) {
	t.Skip("unimplemented: SentenceRangeTest.testSpecialCase")
}
func TestSentenceRange_ExtraWhitespaceCase(t *testing.T) {
	// Port of ExtraWhitespaceCase using GetRangesFromSentences (no full check2 pipeline).
	text := "Hello, how are you?     This is an test."
	annotated := markup.NewAnnotatedTextBuilder().AddText(text).Build()
	sentences := []string{"Hello, how are you?     ", "This is an test."}
	ranges := GetRangesFromSentences(annotated, sentences)
	require.Len(t, ranges, 2)
	require.Equal(t, 0, ranges[0].GetFromPos())
	require.Equal(t, 19, ranges[0].GetToPos())
	require.Equal(t, "Hello, how are you?", text[ranges[0].GetFromPos():ranges[0].GetToPos()])
	require.Equal(t, 24, ranges[1].GetFromPos())
	require.Equal(t, 40, ranges[1].GetToPos())
	require.Equal(t, "This is an test.", text[ranges[1].GetFromPos():ranges[1].GetToPos()])
}
