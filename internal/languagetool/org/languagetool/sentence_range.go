package languagetool

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/markup"
)

// SentenceRange ports org.languagetool.SentenceRange.
type SentenceRange struct {
	FromPos int
	ToPos   int
}

func NewSentenceRange(fromPos, toPos int) SentenceRange {
	return SentenceRange{FromPos: fromPos, ToPos: toPos}
}

func (r SentenceRange) GetFromPos() int { return r.FromPos }
func (r SentenceRange) GetToPos() int   { return r.ToPos }
func (r SentenceRange) String() string  { return fmt.Sprintf("%d-%d", r.FromPos, r.ToPos) }

var (
	beginsWithSpace = regexp.MustCompile(`^\s*`)
	endsWithSpace   = regexp.MustCompile(`\s+$`)
)

// GetRangesFromSentences ports SentenceRange.getRangesFromSentences.
// Positions are UTF-16 code unit offsets (Java String).
func GetRangesFromSentences(annotatedText *markup.AnnotatedText, sentences []string) []SentenceRange {
	var sentenceRanges []SentenceRange
	pos := 0
	markupTextLength := utf16LenSR(annotatedText.GetTextWithMarkup())
	diff := markupTextLength - utf16LenSR(annotatedText.GetPlainText())
	for _, sentence := range sentences {
		if strings.TrimSpace(sentence) == "" {
			// No content no sentence
			pos += utf16LenSR(sentence)
			continue
		}
		sentenceNoBeginWhitespace := beginsWithSpace.ReplaceAllString(sentence, "")
		sentenceNoEndWhitespace := endsWithSpace.ReplaceAllString(sentence, "")
		fromPos := pos + (utf16LenSR(sentence) - utf16LenSR(sentenceNoBeginWhitespace))
		toPos := pos + utf16LenSR(sentenceNoEndWhitespace)

		fromPosOrig := fromPos + diff
		toPosOrig := toPos + diff
		if fromPosOrig != markupTextLength {
			fromPosOrig = annotatedText.GetOriginalTextPositionFor(fromPos, false)
		}
		if toPosOrig != markupTextLength {
			toPosOrig = annotatedText.GetOriginalTextPositionFor(toPos, true)
		}
		sentenceRanges = append(sentenceRanges, NewSentenceRange(fromPosOrig, toPosOrig))
		pos += utf16LenSR(sentence)
	}
	return sentenceRanges
}

func utf16LenSR(s string) int {
	return len(utf16.Encode([]rune(s)))
}
