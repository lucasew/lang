package languagetool

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// AnalyzePlain ports a minimal getAnalyzedSentence for demo/rule unit tests:
// SENT_START + WordTokenizer tokens as untagged AnalyzedTokenReadings with start positions.
func AnalyzePlain(text string) *AnalyzedSentence {
	wt := tokenizers.NewWordTokenizer()
	raw := wt.Tokenize(text)
	positions := tokenizers.BuildPositions(raw)
	// tokens: SENT_START at 0, then each raw token
	readings := make([]*AnalyzedTokenReadings, 0, len(raw)+1)
	ss := SentenceStartTagName
	startTok := NewAnalyzedToken("", &ss, nil)
	startR := NewAnalyzedTokenReadings(startTok)
	startR.SetStartPos(0)
	readings = append(readings, startR)
	for i, tok := range raw {
		at := NewAnalyzedToken(tok, nil, nil)
		// whitespaceBefore: if previous is whitespace... simple: false unless after space
		if i > 0 {
			// not setting for first version
		}
		ar := NewAnalyzedTokenReadingsAt(at, positions[i])
		readings = append(readings, ar)
	}
	return NewAnalyzedSentence(readings)
}

// CheckWhitespaceOnly runs MultipleWhitespace-style single-sentence check via callback.
// Kept in languagetool package for test helpers.
func AnalyzeSentences(text string) []*AnalyzedSentence {
	// single sentence for unit tests
	return []*AnalyzedSentence{AnalyzePlain(text)}
}
