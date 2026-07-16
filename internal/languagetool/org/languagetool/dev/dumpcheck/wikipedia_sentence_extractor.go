package dumpcheck

import (
	"io"
	"strings"
	"unicode"
)

// SkipSentence ports WikipediaSentenceExtractor.skipSentence:
// empty or starting with lowercase are skipped.
func SkipSentence(sentence string) bool {
	trim := strings.TrimSpace(sentence)
	if trim == "" {
		return true
	}
	r := []rune(trim)[0]
	return unicode.IsLower(r)
}

// ExtractWikipediaSentences ports WikipediaSentenceExtractor.extract body:
// streams accepted sentences from source to w, one per line.
func ExtractWikipediaSentences(source SentenceSource, w io.Writer) (int, error) {
	n := 0
	for source.HasNext() {
		sent, err := source.Next()
		if err != nil {
			return n, err
		}
		if SkipSentence(sent.GetText()) {
			continue
		}
		if _, err := io.WriteString(w, sent.GetText()); err != nil {
			return n, err
		}
		if _, err := io.WriteString(w, "\n"); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
