package morfologik

// SpellerED ports the Oflazer edit-distance core of morfologik.speller.Speller
// (setWordAndCandidate, ed, cuted) without FSA traversal. Used for HMatrix-backed
// distance calculation and as the foundation for findRepl.
//
// Java Speller fields used: wordProcessed, wordLen, candidate, candLen,
// effectEditDistance, hMatrix, editDistance.
type SpellerED struct {
	editDistance       int
	effectEditDistance int
	hMatrix            *HMatrix
	wordProcessed      []rune
	wordLen            int
	candidate          []rune
	candLen            int
	// areEqual options (Java dictionaryMetadata)
	IgnoreDiacritics bool
	ConvertCase      bool
	EquivalentChars  map[rune][]rune
}

// MaxWordLength ports Speller.MAX_WORD_LENGTH.
const MaxWordLength = 120

// NewSpellerED ports Speller construction of HMatrix(editDistance, MAX_WORD_LENGTH).
func NewSpellerED(editDistance int) *SpellerED {
	if editDistance < 1 {
		editDistance = 1
	}
	return &SpellerED{
		editDistance:       editDistance,
		effectEditDistance: editDistance,
		hMatrix:            NewHMatrix(editDistance, MaxWordLength),
	}
}

// SetWordAndCandidate ports Speller.setWordAndCandidate (test / ed setup).
func (s *SpellerED) SetWordAndCandidate(word, candidate string) {
	if s == nil {
		return
	}
	s.wordProcessed = []rune(word)
	s.wordLen = len(s.wordProcessed)
	s.candidate = []rune(candidate)
	s.candLen = len(s.candidate)
	// Java: effectEditDistance = wordLen <= editDistance ? wordLen - 1 : editDistance
	if s.wordLen <= s.editDistance {
		s.effectEditDistance = s.wordLen - 1
		if s.effectEditDistance < 0 {
			s.effectEditDistance = 0
		}
	} else {
		s.effectEditDistance = s.editDistance
	}
	s.hMatrix.Reset()
}

// GetWordLen ports Speller.getWordLen.
func (s *SpellerED) GetWordLen() int {
	if s == nil {
		return 0
	}
	return s.wordLen
}

// GetCandLen ports Speller.getCandLen.
func (s *SpellerED) GetCandLen() int {
	if s == nil {
		return 0
	}
	return s.candLen
}

// GetEffectiveED ports Speller.getEffectiveED (package-private in Java tests).
func (s *SpellerED) GetEffectiveED() int {
	if s == nil {
		return 0
	}
	return s.effectEditDistance
}

// Ed ports Speller.ed — Damerau-Levenshtein step writing into HMatrix.
// i = length of misspelled-1; j = length of candidate-1; wordIndex/candIndex are last char indices.
func (s *SpellerED) Ed(i, j, wordIndex, candIndex int) int {
	if s == nil || s.hMatrix == nil {
		return 0
	}
	var result int
	if s.areEqual(s.wordProcessed[wordIndex], s.candidate[candIndex]) {
		// last characters are the same
		result = s.hMatrix.Get(i, j)
	} else if wordIndex > 0 && candIndex > 0 &&
		s.wordProcessed[wordIndex] == s.candidate[candIndex-1] &&
		s.wordProcessed[wordIndex-1] == s.candidate[candIndex] {
		// transposition
		a := s.hMatrix.Get(i-1, j-1)
		b := s.hMatrix.Get(i+1, j)
		c := s.hMatrix.Get(i, j+1)
		result = 1 + min3int(a, b, c)
	} else {
		// replace / delete / insert
		a := s.hMatrix.Get(i, j)
		b := s.hMatrix.Get(i+1, j)
		c := s.hMatrix.Get(i, j+1)
		result = 1 + min3int(a, b, c)
	}
	s.hMatrix.Set(i+1, j+1, result)
	return result
}

// Cuted ports Speller.cuted — cut-off edit distance at current depth.
func (s *SpellerED) Cuted(depth, wordIndex, candIndex int) int {
	if s == nil || s.hMatrix == nil {
		return 0
	}
	// l = max(0, depth - effectEditDistance)
	l := depth - s.effectEditDistance
	if l < 0 {
		l = 0
	}
	// u = min(wordLen - 1 - (wordIndex - depth), depth + effectEditDistance)
	u := s.wordLen - 1 - (wordIndex - depth)
	if up := depth + s.effectEditDistance; up < u {
		u = up
	}
	minEd := s.effectEditDistance + 1
	wi := wordIndex + l - depth
	for i := l; i <= u; i++ {
		d := s.Ed(i, depth, wi, candIndex)
		if d < minEd {
			minEd = d
		}
		wi++
	}
	return minEd
}

// GetEditDistance ports SpellerTest.getEditDistance (full word vs candidate).
func (s *SpellerED) GetEditDistance(word, candidate string) int {
	s.SetWordAndCandidate(word, candidate)
	maxDistance := s.GetEffectiveED()
	candidateLen := s.GetCandLen()
	wordLen := s.GetWordLen()
	ed := 0
	for i := 0; i < candidateLen; i++ {
		if s.Cuted(i, i, i) <= maxDistance {
			if absInt(wordLen-1-i) <= maxDistance {
				ed = s.Ed(wordLen-1, i, wordLen-1, i)
			}
		}
	}
	return ed
}

// GetCutOffDistance ports SpellerTest.getCutOffDistance (min cuted along cand extension).
func (s *SpellerED) GetCutOffDistance(word, candidate string) int {
	s.SetWordAndCandidate(word, candidate)
	diff := s.GetCandLen() - s.GetWordLen()
	if diff <= 0 {
		// Java allocates ced[candLen-wordLen]; empty → 0
		// equal length: still useful to report cuted at end
		if s.GetCandLen() == s.GetWordLen() && s.GetWordLen() > 0 {
			// reporter/reporter: walk full and take final ed path
			return s.GetEditDistance(word, candidate)
		}
		return 0
	}
	ced := make([]int, diff)
	for i := 0; i < diff; i++ {
		ced[i] = s.Cuted(s.GetWordLen()+i, s.GetWordLen()+i, s.GetWordLen()+i)
	}
	// min
	minV := ced[0]
	for _, v := range ced[1:] {
		if v < minV {
			minV = v
		}
	}
	return minV
}

func (s *SpellerED) areEqual(x, y rune) bool {
	if x == y {
		return true
	}
	opt := SuggestOpts{
		IgnoreDiacritics:    s.IgnoreDiacritics,
		ConvertCase:         s.ConvertCase,
		EquivalentChars:     s.EquivalentChars,
		SymmetricEquivalent: false, // Java areEqual is one-way only
	}
	return runesEqualUnderOpts(x, y, opt)
}

func min3int(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
