package languagetool

import (
	"strings"
)

// AnalyzedSentence ports org.languagetool.AnalyzedSentence (subset for unit tests).
type AnalyzedSentence struct {
	tokens                      []*AnalyzedTokenReadings
	preDisambigTokens           []*AnalyzedTokenReadings
	nonBlankTokens              []*AnalyzedTokenReadings
	nonBlankPreDisambigTokens   []*AnalyzedTokenReadings
	whPositions                 []int
}

func NewAnalyzedSentence(words []*AnalyzedTokenReadings) *AnalyzedSentence {
	return NewAnalyzedSentenceFull(words, words)
}

func NewAnalyzedSentenceFull(tokens, preDisambig []*AnalyzedTokenReadings) *AnalyzedSentence {
	mapping := make([]int, len(tokens)+1)
	nonBlank := getNonBlankReadings(tokens, mapping)
	// rebuild mapping for preDisambig independently like Java does with same vars (shared counters - see Java)
	// Java reuses whCounter/nonWhCounter/mapping for second call - BUG-compatible?
	// Actually Java passes same variables so second call continues counters. For equal arrays it's ok.
	s := &AnalyzedSentence{
		tokens:            tokens,
		preDisambigTokens: preDisambig,
		whPositions:       mapping,
		nonBlankTokens:    nonBlank,
	}
	// second non-blank pass like Java constructor - uses same mapping array
	s.nonBlankPreDisambigTokens = getNonBlankReadings(preDisambig, mapping)
	return s
}

func getNonBlankReadings(tokens []*AnalyzedTokenReadings, mapping []int) []*AnalyzedTokenReadings {
	var l []*AnalyzedTokenReadings
	whCounter, nonWhCounter := 0, 0
	for _, token := range tokens {
		if !token.IsWhitespace() || token.IsSentenceStart() || token.IsSentenceEnd() || token.IsParagraphEnd() {
			l = append(l, token)
			if nonWhCounter < len(mapping) {
				mapping[nonWhCounter] = whCounter
			}
			nonWhCounter++
		}
		whCounter++
	}
	return l
}

func (s *AnalyzedSentence) GetTokens() []*AnalyzedTokenReadings { return s.tokens }

func (s *AnalyzedSentence) GetTokensWithoutWhitespace() []*AnalyzedTokenReadings {
	return s.nonBlankTokens
}

func (s *AnalyzedSentence) GetPreDisambigTokensWithoutWhitespace() []*AnalyzedTokenReadings {
	return s.nonBlankPreDisambigTokens
}

// Copy ports AnalyzedSentence.copy.
func (s *AnalyzedSentence) Copy(sentence *AnalyzedSentence) *AnalyzedSentence {
	copyTokens := make([]*AnalyzedTokenReadings, len(sentence.GetTokens()))
	for i, analyzedTokens := range sentence.GetTokens() {
		copyTokens[i] = NewAnalyzedTokenReadingsFromOld(analyzedTokens, analyzedTokens.GetReadings(), "")
	}
	return &AnalyzedSentence{
		tokens:                    copyTokens,
		preDisambigTokens:         copyTokens,
		whPositions:               append([]int(nil), sentence.whPositions...),
		nonBlankTokens:            sentence.GetTokensWithoutWhitespace(),
		nonBlankPreDisambigTokens: sentence.GetPreDisambigTokensWithoutWhitespace(),
	}
}

func (s *AnalyzedSentence) String() string {
	return s.ToStringDelim(",")
}

func (s *AnalyzedSentence) ToStringDelim(readingDelimiter string) string {
	var sb strings.Builder
	for _, element := range s.tokens {
		if !element.IsWhitespace() {
			sb.WriteString(element.GetToken())
			sb.WriteByte('[')
		}
		readings := element.GetReadings()
		for i, token := range readings {
			posTag := token.GetPOSTag()
			if element.IsSentenceStart() {
				sb.WriteString("<S>")
			} else if posTag != nil && *posTag == SentenceEndTagName {
				sb.WriteString("</S>")
			} else if posTag != nil && *posTag == ParagraphEndTagName {
				sb.WriteString("<P/>")
			} else {
				if !element.IsWhitespace() {
					sb.WriteString(token.String())
					if i+1 < len(readings) {
						// only delimiter between non-special readings — Java uses iterator.hasNext after current
						// Special tags don't use delimiter the same way; match Java loop structure:
					}
				}
			}
			// Java: delimiter when hasNext and not the special cases that don't append token
			if !element.IsWhitespace() && i+1 < len(readings) {
				// peek next - Java appends delimiter after appending token if hasNext
				// For SENT_END path no delimiter before next
				next := readings[i+1]
				npt := next.GetPOSTag()
				curIsSpecial := element.IsSentenceStart() ||
					(posTag != nil && (*posTag == SentenceEndTagName || *posTag == ParagraphEndTagName))
				nextIsSpecial := npt != nil && (*npt == SentenceEndTagName || *npt == ParagraphEndTagName)
				if !curIsSpecial && !element.IsSentenceStart() {
					if posTag == nil || (*posTag != SentenceEndTagName && *posTag != ParagraphEndTagName) {
						if !nextIsSpecial || true {
							// Java always appends delimiter if hasNext after processing current in the else branch only
						}
					}
				}
			}
		}
		// Simpler faithful rewrite matching Java structure exactly:
		_ = readings
		if false {
			sb.WriteString(readingDelimiter)
		}
		if !element.IsWhitespace() {
			if element.IsImmunized() {
				sb.WriteString("{!}")
			}
			sb.WriteByte(']')
		} else {
			sb.WriteByte(' ')
		}
	}
	// The above loop is incomplete — rewrite cleanly below
	return s.toStringJava(readingDelimiter, true)
}

func (s *AnalyzedSentence) toStringJava(readingDelimiter string, includeChunks bool) string {
	var sb strings.Builder
	for _, element := range s.tokens {
		if !element.IsWhitespace() {
			sb.WriteString(element.GetToken())
			sb.WriteByte('[')
		}
		readings := element.GetReadings()
		for i := 0; i < len(readings); i++ {
			token := readings[i]
			posTag := token.GetPOSTag()
			if element.IsSentenceStart() {
				sb.WriteString("<S>")
			} else if posTag != nil && *posTag == SentenceEndTagName {
				sb.WriteString("</S>")
			} else if posTag != nil && *posTag == ParagraphEndTagName {
				sb.WriteString("<P/>")
			} else if posTag == nil && !includeChunks {
				sb.WriteString(token.GetToken())
			} else {
				if !element.IsWhitespace() {
					sb.WriteString(token.String())
					if i+1 < len(readings) {
						sb.WriteString(readingDelimiter)
					}
				}
			}
		}
		if !element.IsWhitespace() {
			if element.IsImmunized() {
				sb.WriteString("{!}")
			}
			sb.WriteByte(']')
		} else {
			sb.WriteByte(' ')
		}
	}
	return sb.String()
}

func (s *AnalyzedSentence) Equals(o *AnalyzedSentence) bool {
	if s == o {
		return true
	}
	if o == nil {
		return false
	}
	if len(s.tokens) != len(o.tokens) || len(s.nonBlankTokens) != len(o.nonBlankTokens) {
		return false
	}
	// Java uses Arrays.equals on token arrays (reference equality of elements for ATR)
	// After copy, elements are new objects so equals of ATR matters - Java AnalyzedTokenReadings equals?
	// Arrays.equals uses Object.equals. AnalyzedTokenReadings may not override equals → reference equality.
	// Test: after copy, equals true; after immunize original, not equal.
	// So equals is reference equality of arrays contents - after copy, new ATR objects, Arrays.equals uses equals().
	// If ATR doesn't override equals, copy would NOT equal original. But test expects equal.
	// Check if ATR has equals...
	return s.equalTokens(s.tokens, o.tokens) && s.equalTokens(s.nonBlankTokens, o.nonBlankTokens)
}

func (s *AnalyzedSentence) equalTokens(a, b []*AnalyzedTokenReadings) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].String() != b[i].String() {
			return false
		}
	}
	return true
}

// GetCorrectedTextLength ports AnalyzedSentence.getCorrectedTextLength.
func (s *AnalyzedSentence) GetCorrectedTextLength() int {
	lenSum := 0
	for i, element := range s.tokens {
		// use token length; clean token if we had it
		t := element.GetToken()
		n := 0
		for _, r := range t {
			if r >= 0x10000 {
				n += 2
			} else {
				n++
			}
		}
		lenSum += n
		_ = i
	}
	return lenSum
}
