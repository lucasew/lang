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
	// Port of SentenceRangeTest.testGermanSentenceRange without full check2:
	// feed pre-tokenized sentences (as SRX would emit) into GetRangesFromSentences.
	contents := []string{
		"LanguageTool",
		"Unsere Grammatik-, Stil- und Rechtschreibprüfung ist in vielen Sprachen verfügbar und wird von Millionen Menschen weltweit genutzt",
		"Probieren Sie den LanguageTool-Editor aus.",
		"Bekommen Sie Tipps zur Verbesserung Ihrer Rechtschreibung (inklusive Kommasetzung u.v.m.) während Sie E-Mails schreiben, bloggen oder einfach nur twittern.",
		"LanguageTool erkennt automatisch, in welcher Sprache Sie schreiben.",
		"Um Ihre Daten zu schützen, werden vom Browser-Add-on keine Texte gespeichert.",
		"Holen Sie alles aus Ihren Dokumenten heraus und liefern Sie fehlerfreie Ergebnisse ab.",
		"Egal, ob Sie an einer Dissertation arbeiten, einen Aufsatz oder ein Buch schreiben oder einfach nur Notizen machen.",
		"\uFEFF\u2063",
		"Professionalisieren Sie die Kommunikation Ihres Teams mit der Grammatik- und Stilprüfung von LanguageTool.",
		"Voll unterstützt (Rechtschreibung, Grammatik- und Stilhinweise):",
		"Englisch",
		"Deutsch",
		"Französisch",
		"Spanisch",
		"Niederländisch",
		"Danke, dass Sie es ausprobieren!",
	}
	sentences := make([]string, len(contents))
	for i, c := range contents {
		sentences[i] = "\n\n" + c + "\n\n"
	}
	text := strings.Join(sentences, "")
	annotated := markup.NewAnnotatedTextBuilder().AddText(text).Build()
	ranges := GetRangesFromSentences(annotated, sentences)
	require.Len(t, ranges, 17)
	require.Equal(t, "LanguageTool", utf16SubstrSR(text, ranges[0].GetFromPos(), ranges[0].GetToPos()))
	require.Equal(t, contents[1], utf16SubstrSR(text, ranges[1].GetFromPos(), ranges[1].GetToPos()))
	require.Equal(t, contents[2], utf16SubstrSR(text, ranges[2].GetFromPos(), ranges[2].GetToPos()))
	require.Equal(t, contents[3], utf16SubstrSR(text, ranges[3].GetFromPos(), ranges[3].GetToPos()))
	require.Equal(t, contents[4], utf16SubstrSR(text, ranges[4].GetFromPos(), ranges[4].GetToPos()))
	require.Equal(t, contents[5], utf16SubstrSR(text, ranges[5].GetFromPos(), ranges[5].GetToPos()))
	require.Equal(t, contents[6], utf16SubstrSR(text, ranges[6].GetFromPos(), ranges[6].GetToPos()))
	require.Equal(t, contents[7], utf16SubstrSR(text, ranges[7].GetFromPos(), ranges[7].GetToPos()))
	require.Equal(t, contents[9], utf16SubstrSR(text, ranges[9].GetFromPos(), ranges[9].GetToPos()))
	require.Equal(t, contents[10], utf16SubstrSR(text, ranges[10].GetFromPos(), ranges[10].GetToPos()))
	require.Equal(t, "Englisch", utf16SubstrSR(text, ranges[11].GetFromPos(), ranges[11].GetToPos()))
	require.Equal(t, "Deutsch", utf16SubstrSR(text, ranges[12].GetFromPos(), ranges[12].GetToPos()))
	require.Equal(t, "Französisch", utf16SubstrSR(text, ranges[13].GetFromPos(), ranges[13].GetToPos()))
	require.Equal(t, "Spanisch", utf16SubstrSR(text, ranges[14].GetFromPos(), ranges[14].GetToPos()))
	require.Equal(t, "Niederländisch", utf16SubstrSR(text, ranges[15].GetFromPos(), ranges[15].GetToPos()))
	require.Equal(t, "Danke, dass Sie es ausprobieren!", utf16SubstrSR(text, ranges[16].GetFromPos(), ranges[16].GetToPos()))
}
func TestSentenceRange_EnglishSentenceRange(t *testing.T) {
	// Port of SentenceRangeTest.testEnglishSentenceRange (pre-tokenized path).
	contents := []string{
		"LanguageTool",
		"LanguageTool’s multilingual grammar, style, and spell checker is used by millions of people around the world.",
		"Trusted by our partners and customers",
		"Receive tips on how to improve your text (including punctuation advice etc.) while typing an e-mail, a blog post or just a simple tweet.",
		"Whatever language you're using, LanguageTool will automatically detect it and provide suggestions.",
		"To respect your privacy, no text is stored by the browser add-on.",
		"Get the best out of your docs and deliver error-free results.",
		"No matter whether you're working on a dissertation, an essay, or a book, or you just want to note down something.",
		"\uFEFF\u2063",
		"Professionalize your team's communication with LanguageTool's grammar and style checker.",
		"Fully supported (spelling, grammar, style hints):",
		"English",
		"German",
		"French",
		"Spanish",
		"Dutch",
		"Thanks for checking it out!",
	}
	sentences := make([]string, len(contents))
	for i, c := range contents {
		sentences[i] = "\n\n" + c + "\n\n"
	}
	text := strings.Join(sentences, "")
	annotated := markup.NewAnnotatedTextBuilder().AddText(text).Build()
	ranges := GetRangesFromSentences(annotated, sentences)
	require.Len(t, ranges, 17)
	require.Equal(t, "LanguageTool", utf16SubstrSR(text, ranges[0].GetFromPos(), ranges[0].GetToPos()))
	require.Equal(t, contents[1], utf16SubstrSR(text, ranges[1].GetFromPos(), ranges[1].GetToPos()))
	require.Equal(t, contents[2], utf16SubstrSR(text, ranges[2].GetFromPos(), ranges[2].GetToPos()))
	require.Equal(t, contents[3], utf16SubstrSR(text, ranges[3].GetFromPos(), ranges[3].GetToPos()))
	require.Equal(t, contents[4], utf16SubstrSR(text, ranges[4].GetFromPos(), ranges[4].GetToPos()))
	require.Equal(t, contents[5], utf16SubstrSR(text, ranges[5].GetFromPos(), ranges[5].GetToPos()))
	require.Equal(t, contents[6], utf16SubstrSR(text, ranges[6].GetFromPos(), ranges[6].GetToPos()))
	require.Equal(t, contents[7], utf16SubstrSR(text, ranges[7].GetFromPos(), ranges[7].GetToPos()))
	require.Equal(t, contents[9], utf16SubstrSR(text, ranges[9].GetFromPos(), ranges[9].GetToPos()))
	require.Equal(t, contents[10], utf16SubstrSR(text, ranges[10].GetFromPos(), ranges[10].GetToPos()))
	require.Equal(t, "English", utf16SubstrSR(text, ranges[11].GetFromPos(), ranges[11].GetToPos()))
	require.Equal(t, "German", utf16SubstrSR(text, ranges[12].GetFromPos(), ranges[12].GetToPos()))
	require.Equal(t, "French", utf16SubstrSR(text, ranges[13].GetFromPos(), ranges[13].GetToPos()))
	require.Equal(t, "Spanish", utf16SubstrSR(text, ranges[14].GetFromPos(), ranges[14].GetToPos()))
	require.Equal(t, "Dutch", utf16SubstrSR(text, ranges[15].GetFromPos(), ranges[15].GetToPos()))
	require.Equal(t, "Thanks for checking it out!", utf16SubstrSR(text, ranges[16].GetFromPos(), ranges[16].GetToPos()))
}
func TestSentenceRange_SpecialCase(t *testing.T) {
	// Special characters / BOM-like content still produce ranges for non-empty sentences.
	sentences := []string{"\uFEFFHello.", " World."}
	text := strings.Join(sentences, "")
	annotated := markup.NewAnnotatedTextBuilder().AddText(text).Build()
	ranges := GetRangesFromSentences(annotated, sentences)
	require.Len(t, ranges, 2)
	require.Greater(t, ranges[0].GetToPos(), ranges[0].GetFromPos())
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
