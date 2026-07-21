package implementation

import (
	"math"
	"strings"
	"unicode"
	"unicode/utf16"
)

// corpusWordCountN ports SymSpell.N — total word count constant used for Naive Bayes log probs.
// Java: private static long N = 1024908267229L;
const corpusWordCountN int64 = 1024908267229

// SegmentedSuggestion ports SymSpell.SegmentedSuggestion.
type SegmentedSuggestion struct {
	SegmentedString   string
	CorrectedString   string
	DistanceSum       int
	ProbabilityLogSum float64
}

// WordSegmentation ports wordSegmentation(input) using maxDictionaryEditDistance and maxLength.
func (s *SymSpell) WordSegmentation(input string) SegmentedSuggestion {
	if s == nil {
		return SegmentedSuggestion{}
	}
	return s.WordSegmentationMax(input, s.maxDictionaryEditDistance, s.maxLength)
}

// WordSegmentationEdit ports wordSegmentation(input, maxEditDistance).
func (s *SymSpell) WordSegmentationEdit(input string, maxEditDistance int) SegmentedSuggestion {
	if s == nil {
		return SegmentedSuggestion{}
	}
	return s.WordSegmentationMax(input, maxEditDistance, s.maxLength)
}

// WordSegmentationMax ports wordSegmentation(input, maxEditDistance, maxSegmentationWordLength).
// All substring/length/charAt operations use Java UTF-16 units.
func (s *SymSpell) WordSegmentationMax(input string, maxEditDistance, maxSegmentationWordLength int) SegmentedSuggestion {
	if s == nil || input == "" {
		return SegmentedSuggestion{}
	}
	inputLen := javaStringLen(input)
	arraySize := maxSegmentationWordLength
	if inputLen < arraySize {
		arraySize = inputLen
	}
	if arraySize <= 0 {
		return SegmentedSuggestion{}
	}
	compositions := make([]SegmentedSuggestion, arraySize)
	circularIndex := -1

	// outer loop (column): all possible part start positions
	for j := 0; j < inputLen; j++ {
		// inner loop (row): all possible part lengths
		imax := inputLen - j
		if maxSegmentationWordLength < imax {
			imax = maxSegmentationWordLength
		}
		for i := 1; i <= imax; i++ {
			part := javaSubstring(input, j, j+i)
			separatorLength := 0
			topEd := 0
			var topProbabilityLog float64
			var topResult string

			// Java Character.isWhitespace(part.charAt(0))
			if javaStringLen(part) > 0 && isJavaWhitespaceChar(part) {
				// remove space for levenshtein calculation
				part = javaSubstring(part, 1, javaStringLen(part))
			} else {
				// add ed+1: space did not exist, had to be inserted
				separatorLength = 1
			}

			// remove space from part, add number of removed spaces to topEd
			topEd += javaStringLen(part)
			// remove space (ASCII space only — Java part.replace(" ", ""))
			part = strings.ReplaceAll(part, " ", "")
			topEd -= javaStringLen(part)

			results := s.LookupMax(part, VerbosityTop, maxEditDistance)
			if len(results) > 0 {
				topResult = results[0].Term
				topEd += results[0].Distance
				// Naive Bayes: log10(count / N)
				topProbabilityLog = math.Log10(float64(results[0].Count) / float64(corpusWordCountN))
			} else {
				topResult = part
				// default, if word not found
				topEd += javaStringLen(part)
				// log10(10.0 / (N * 10^len))
				topProbabilityLog = math.Log10(10.0 / (float64(corpusWordCountN) * math.Pow(10.0, float64(javaStringLen(part)))))
			}

			destinationIndex := (i + circularIndex) % arraySize
			// Go % with negative: circularIndex starts -1, i>=1 → can be 0.. 
			// Java: ((i + circularIndex) % arraySize) — when circularIndex=-1, i=1 → 0.
			// In Go, (-1+1)%n = 0; but (i + (-1)) when i=1 is 0. Good.
			// When circularIndex=-1 and i=arraySize: (arraySize-1)%arraySize ok.
			if destinationIndex < 0 {
				destinationIndex += arraySize
			}

			if j == 0 {
				// set values in first loop
				compositions[destinationIndex].SegmentedString = part
				compositions[destinationIndex].CorrectedString = topResult
				compositions[destinationIndex].DistanceSum = topEd
				compositions[destinationIndex].ProbabilityLogSum = topProbabilityLog
			} else if i == maxSegmentationWordLength ||
				// replace if better probabilityLogSum at same ed or one space difference
				(((compositions[circularIndex].DistanceSum+topEd == compositions[destinationIndex].DistanceSum) ||
					(compositions[circularIndex].DistanceSum+separatorLength+topEd == compositions[destinationIndex].DistanceSum)) &&
					(compositions[destinationIndex].ProbabilityLogSum < compositions[circularIndex].ProbabilityLogSum+topProbabilityLog)) ||
				// replace if smaller edit distance
				(compositions[circularIndex].DistanceSum+separatorLength+topEd < compositions[destinationIndex].DistanceSum) {
				compositions[destinationIndex].SegmentedString = compositions[circularIndex].SegmentedString + " " + part
				compositions[destinationIndex].CorrectedString = compositions[circularIndex].CorrectedString + " " + topResult
				compositions[destinationIndex].DistanceSum = compositions[circularIndex].DistanceSum + topEd
				compositions[destinationIndex].ProbabilityLogSum = compositions[circularIndex].ProbabilityLogSum + topProbabilityLog
			}
		}
		circularIndex++
		if circularIndex >= arraySize {
			circularIndex = 0
		}
	}
	if circularIndex < 0 || circularIndex >= arraySize {
		return SegmentedSuggestion{}
	}
	return compositions[circularIndex]
}

// isJavaWhitespaceChar ports Character.isWhitespace(part.charAt(0)) for first UTF-16 unit.
// Java Character.isWhitespace covers a fixed set; for BMP we use unicode.IsSpace on the first rune
// when it is not a surrogate pair start, else treat high surrogate as non-whitespace (rare for WS).
func isJavaWhitespaceChar(part string) bool {
	u := utf16.Encode([]rune(part))
	if len(u) == 0 {
		return false
	}
	// Reconstruct first code point if surrogate pair
	r := utf16.Decode(u[:1])
	if len(r) == 0 {
		return false
	}
	// If high surrogate alone, Decode may produce replacement — treat as not whitespace
	ch := r[0]
	// Java Character.isWhitespace(char) for BMP code units.
	// For simplicity mirror unicode.IsSpace for common whitespace used by SymSpell inputs.
	return unicode.IsSpace(ch)
}
