package languagetool

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// AnalyzePlain ports a minimal getAnalyzedSentence for demo/rule unit tests:
// SENT_START + WordTokenizer tokens as untagged AnalyzedTokenReadings with start positions.
// Sets whitespaceBefore from the previous raw token (JLanguageTool analysis loop).
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
	prev := ""
	for i, tok := range raw {
		at := NewAnalyzedToken(tok, nil, nil)
		ar := NewAnalyzedTokenReadingsAt(at, positions[i])
		if prev != "" {
			ar.SetWhitespaceBeforeToken(prev)
		}
		readings = append(readings, ar)
		prev = tok
	}
	return NewAnalyzedSentence(readings)
}

// CheckWhitespaceOnly runs MultipleWhitespace-style single-sentence check via callback.
// Kept in languagetool package for test helpers.
func AnalyzeSentences(text string) []*AnalyzedSentence {
	// single sentence for unit tests
	return []*AnalyzedSentence{AnalyzePlain(text)}
}

// SplitAndAnalyze splits on .!? boundaries for SentenceWhitespaceRule unit tests.
// Trailing single space after terminator is attached to the previous sentence
// (so prevSentenceEndsWithWhitespace matches LT SRX-ish behavior for these tests).
// Periods inside URLs/domains (e.g. example.com) are not treated as boundaries.
func SplitAndAnalyze(text string) []*AnalyzedSentence {
	if text == "" {
		return nil
	}
	var parts []string
	start := 0
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '.' || r == '!' || r == '?' {
			// Do not split on '.' when next is lowercase letter or digit (domain/decimal).
			// Uppercase after '.' without space is still a sentence boundary
			// (SentenceWhitespaceRule "text.And next").
			if r == '.' && i+1 < len(runes) {
				n := runes[i+1]
				if (n >= 'a' && n <= 'z') || (n >= '0' && n <= '9') {
					continue
				}
			}
			end := i + 1
			// include following single space/newline as part of this sentence
			if end < len(runes) && (runes[end] == ' ' || runes[end] == '\n' || runes[end] == '\u00A0') {
				// only one whitespace for "ends with whitespace" check
				if runes[end] == '\n' && end+1 < len(runes) && runes[end+1] == '\n' {
					// paragraph break: include first newline only? good tests have \n between sentences
					end++
					// if double newline, include second as well for "\n\n" good case
					if end < len(runes) && runes[end] == '\n' {
						end++
					}
				} else if runes[end] == ' ' || runes[end] == '\u00A0' {
					end++
				} else if runes[end] == '\n' {
					end++
				}
			}
			parts = append(parts, string(runes[start:end]))
			start = end
			i = end - 1
		}
	}
	if start < len(runes) {
		parts = append(parts, string(runes[start:]))
	}
	out := make([]*AnalyzedSentence, 0, len(parts))
	offset := 0
	for _, p := range parts {
		if p == "" {
			continue
		}
		s := AnalyzePlain(p)
		// shift token positions by offset for multi-sentence
		if offset > 0 {
			shiftSentence(s, offset)
		}
		out = append(out, s)
		// offset by UTF-16 length of part
		for _, r := range p {
			if r >= 0x10000 {
				offset += 2
			} else {
				offset++
			}
		}
	}
	return out
}

func shiftSentence(s *AnalyzedSentence, delta int) {
	for _, t := range s.GetTokens() {
		t.SetStartPos(t.GetStartPos() + delta)
	}
}


// AnalyzeTextDemo splits text into sentences for Demo-like unit tests.
// Paragraph boundaries: blank lines (\n\n). Sentence-local token positions
// (as LT does); TextLevelRule.match accumulates pos across sentences.
func AnalyzeTextDemo(text string) []*AnalyzedSentence {
	paras := strings.Split(text, "\n\n")
	var out []*AnalyzedSentence
	for pi, para := range paras {
		chunk := para
		var sents []*AnalyzedSentence
		if strings.Contains(chunk, ". ") || strings.Contains(chunk, ".\n") || strings.Contains(chunk, "! ") || strings.Contains(chunk, "? ") {
			sents = SplitAndAnalyze(chunk)
		} else if chunk != "" {
			sents = []*AnalyzedSentence{AnalyzePlain(chunk)}
		}
		if pi < len(paras)-1 && len(sents) > 0 {
			// Ensure last sentence of paragraph ends with \n\n for isParagraphEnd
			if len(sents) == 1 {
				sents = []*AnalyzedSentence{AnalyzePlain(chunk + "\n\n")}
			} else {
				sents = SplitAndAnalyze(chunk + "\n\n")
			}
		}
		out = append(out, sents...)
	}
	if len(out) == 0 && text != "" {
		return []*AnalyzedSentence{AnalyzePlain(text)}
	}
	return out
}
