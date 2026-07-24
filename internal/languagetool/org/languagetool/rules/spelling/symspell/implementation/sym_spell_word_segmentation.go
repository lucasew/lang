package implementation

import (
	"math"
	"strings"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
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
	// Java: if (input.isEmpty()) return new SegmentedSuggestion();
	if s == nil || input == "" {
		return SegmentedSuggestion{}
	}
	inputLen := javaStringLen(input)
	// Java: int arraySize = Math.min(maxSegmentationWordLength, input.length());
	arraySize := maxSegmentationWordLength
	if inputLen < arraySize {
		arraySize = inputLen
	}
	compositions := make([]SegmentedSuggestion, arraySize)
	// Java: for (int i = 0; i < arraySize; i++) compositions[i] = new SegmentedSuggestion();
	// Go zero values already match empty SegmentedSuggestion defaults.

	circularIndex := -1

	// outer loop (column): all possible part start positions
	for j := 0; j < inputLen; j++ {
		// inner loop (row): all possible part lengths (from start position)
		// part can't be bigger than longest word in dictionary
		imax := inputLen - j
		if maxSegmentationWordLength < imax {
			imax = maxSegmentationWordLength
		}
		for i := 1; i <= imax; i++ {
			// Java: String part = input.substring(j, j + i);
			part := javaSubstring(input, j, j+i)
			separatorLength := 0
			topEd := 0
			var topProbabilityLog float64
			var topResult string

			// Java: Character.isWhitespace(part.charAt(0))
			if isJavaWhitespaceChar(part) {
				// remove space for levenshtein calculation
				part = javaSubstring(part, 1, javaStringLen(part))
			} else {
				// add ed+1: space did not exist, had to be inserted
				separatorLength = 1
			}

			// remove space from part1, add number of removed spaces to topEd
			topEd += javaStringLen(part)
			// Java: part = part.replace(" ", ""); // ASCII space only
			part = strings.ReplaceAll(part, " ", "")
			// add number of removed spaces to ed
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

			// Java: int destinationIndex = ((i + circularIndex) % arraySize);
			// When circularIndex == -1: (i - 1) % arraySize; Go % on non-negative is fine.
			destinationIndex := (i + circularIndex) % arraySize
			if destinationIndex < 0 {
				destinationIndex += arraySize
			}

			// set values in first loop
			if j == 0 {
				compositions[destinationIndex].SegmentedString = part
				compositions[destinationIndex].CorrectedString = topResult
				compositions[destinationIndex].DistanceSum = topEd
				compositions[destinationIndex].ProbabilityLogSum = topProbabilityLog
			} else if (i == maxSegmentationWordLength) ||
				// replace values if better probabilityLogSum, if same edit distance OR one space difference
				((((compositions[circularIndex].DistanceSum+topEd == compositions[destinationIndex].DistanceSum) ||
					(compositions[circularIndex].DistanceSum+separatorLength+topEd == compositions[destinationIndex].DistanceSum)) &&
					(compositions[destinationIndex].ProbabilityLogSum < compositions[circularIndex].ProbabilityLogSum+topProbabilityLog)) ||
					// replace values if smaller edit distance
					(compositions[circularIndex].DistanceSum+separatorLength+topEd < compositions[destinationIndex].DistanceSum)) {
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
	// Java: return compositions[circularIndex];
	return compositions[circularIndex]
}

// isJavaWhitespaceChar ports Character.isWhitespace(part.charAt(0)) for first UTF-16 unit.
// Surrogate code units are never whitespace under Java Character.isWhitespace(char).
func isJavaWhitespaceChar(part string) bool {
	u := utf16.Encode([]rune(part))
	if len(u) == 0 {
		return false
	}
	// Java charAt(0): single UTF-16 code unit (not decoded surrogate pair).
	return tools.CharacterIsWhitespace(rune(u[0]))
}
