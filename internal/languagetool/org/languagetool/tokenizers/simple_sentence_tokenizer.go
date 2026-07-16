package tokenizers

import "unicode"

// SimpleSentenceTokenizer ports Default rules from segment-simple.srx:
// break after [.!?…] followed by whitespace, or [.!?…] followed by uppercase.
type SimpleSentenceTokenizer struct{}

func NewSimpleSentenceTokenizer() *SimpleSentenceTokenizer {
	return &SimpleSentenceTokenizer{}
}

// Tokenize returns sentence segments that concatenate back to text.
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
		// case 1: punct + whitespace → break after one whitespace
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
		// case 2: punct + uppercase → break before uppercase
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
