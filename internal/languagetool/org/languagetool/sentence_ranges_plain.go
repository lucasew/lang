package languagetool

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// PlainSentenceRanges returns byte offsets of SRX sentences in plain text.
// Soft stand-in for SentenceRange on unannotated documents.
func PlainSentenceRanges(text, languageCode string) []SentenceRange {
	if text == "" {
		return nil
	}
	st := tokenizers.NewSRXSentenceTokenizer(languageCode)
	parts := st.Tokenize(text)
	if len(parts) == 0 {
		return []SentenceRange{NewSentenceRange(0, len(text))}
	}
	var out []SentenceRange
	// Map each sentence string back into text by sequential search (same as Check).
	srcRunes := []rune(text)
	searchFrom := 0
	for _, p := range parts {
		if p == "" {
			continue
		}
		pr := []rune(p)
		docBase := indexRunesFrom(srcRunes, pr, searchFrom)
		if docBase < 0 {
			docBase = searchFrom
		}
		// convert rune indices to byte offsets
		fromByte := runeOffsetToByte(text, docBase)
		toByte := runeOffsetToByte(text, docBase+len(pr))
		out = append(out, NewSentenceRange(fromByte, toByte))
		searchFrom = docBase + len(pr)
	}
	return out
}
