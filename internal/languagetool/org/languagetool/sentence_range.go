package languagetool

import (
	"fmt"
	"regexp"
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

// Equal ports SentenceRange.equals (fromPos + toPos only).
func (r SentenceRange) Equal(o SentenceRange) bool {
	return r.FromPos == o.FromPos && r.ToPos == o.ToPos
}

// CompareTo ports SentenceRange.compareTo (order by fromPos).
func (r SentenceRange) CompareTo(o SentenceRange) int {
	if r.FromPos < o.FromPos {
		return -1
	}
	if r.FromPos > o.FromPos {
		return 1
	}
	return 0
}

var (
	// Java Pattern \s without UNICODE_CHARACTER_CLASS: [ \t\n\x0B\f\r]
	// (not Go RE2 \s which includes NBSP and other Unicode spaces).
	beginsWithSpace = regexp.MustCompile(`^[ \t\n\v\f\r]*`)
	endsWithSpace   = regexp.MustCompile(`[ \t\n\v\f\r]+$`)
)

// GetRangesFromSentences ports SentenceRange.getRangesFromSentences.
// Positions are UTF-16 code unit offsets (Java String).
func GetRangesFromSentences(annotatedText *markup.AnnotatedText, sentences []string) []SentenceRange {
	var sentenceRanges []SentenceRange
	pos := 0
	markupTextLength := utf16LenSR(annotatedText.GetTextWithMarkup())
	diff := markupTextLength - utf16LenSR(annotatedText.GetPlainText())
	for _, sentence := range sentences {
		// Java: sentence.trim().isEmpty() — String.trim, not Unicode TrimSpace.
		if javaTrim(sentence) == "" {
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
