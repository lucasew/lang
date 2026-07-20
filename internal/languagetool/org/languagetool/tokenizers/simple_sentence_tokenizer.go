package tokenizers

import (
	"unicode"
)

// SimpleSentenceTokenizer ports org.languagetool.tokenizers.SimpleSentenceTokenizer.
//
// Java extends SRXSentenceTokenizer with "/org/languagetool/tokenizers/segment-simple.srx"
// and AnyLanguage shortCode "xx". The Default languagerule in that file is only:
//
//	break after [.!?…] followed by whitespace (\s)
//	break after [.!?…] followed by uppercase (\p{Lu})
//
// No invent abbreviation / ordinal no-break lists (those live only in full segment.srx
// via language-specific SRXSentenceTokenizer).
type SimpleSentenceTokenizer struct{}

func NewSimpleSentenceTokenizer() *SimpleSentenceTokenizer {
	return &SimpleSentenceTokenizer{}
}

// Tokenize returns sentence segments that concatenate back to text.
// Implements segment-simple.srx Default rules only (Java SimpleSentenceTokenizer).
func (t *SimpleSentenceTokenizer) Tokenize(text string) []string {
	if text == "" {
		return nil
	}
	var out []string
	start := 0
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r != '.' && r != '!' && r != '?' && r != '…' {
			continue
		}
		// consume run of sentence-ending punctuation
		j := i
		for j+1 < len(runes) {
			n := runes[j+1]
			if n == '.' || n == '!' || n == '?' || n == '…' {
				j++
				continue
			}
			break
		}
		// segment-simple.srx: beforebreak [\.!?…]\s → break after one whitespace
		if j+1 < len(runes) && unicode.IsSpace(runes[j+1]) {
			end := j + 2 // include one whitespace (SRX \s)
			if end > len(runes) {
				end = len(runes)
			}
			out = append(out, string(runes[start:end]))
			start = end
			i = end - 1
			continue
		}
		// segment-simple.srx: beforebreak [\.!?…]\p{Lu} → break before uppercase
		if j+1 < len(runes) && unicode.IsUpper(runes[j+1]) {
			end := j + 1
			out = append(out, string(runes[start:end]))
			start = end
			i = end - 1
			continue
		}
		i = j
	}
	if start < len(runes) {
		out = append(out, string(runes[start:]))
	}
	return out
}
