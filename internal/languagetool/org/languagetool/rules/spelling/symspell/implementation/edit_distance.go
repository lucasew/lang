package implementation

// DistanceAlgorithm selects the edit distance algorithm.
type DistanceAlgorithm int

const (
	// Damerau is the only algorithm currently supported (Java DistanceAlgorithm.Damerau).
	Damerau DistanceAlgorithm = iota
)

// EditDistance ports org.languagetool.rules.spelling.symspell.implementation.EditDistance.
type EditDistance struct {
	baseString string
	algorithm  DistanceAlgorithm
	v0, v2     []int
}

func NewEditDistance(baseString string, algorithm DistanceAlgorithm) *EditDistance {
	e := &EditDistance{baseString: baseString, algorithm: algorithm}
	if baseString == "" {
		e.baseString = ""
		return e
	}
	if algorithm == Damerau {
		// Java: new int[baseString.length()] — UTF-16 units; for BMP LT tags/tokens
		// runes match Java char counts. Size by runes (not UTF-8 bytes).
		n := len([]rune(baseString))
		e.v0 = make([]int, n)
		e.v2 = make([]int, n)
	}
	return e
}

// Compare returns edit distance or -1 if greater than maxDistance.
func (e *EditDistance) Compare(string2 string, maxDistance int) int {
	if e == nil {
		return -1
	}
	switch e.algorithm {
	case Damerau:
		return e.DamerauLevenshteinDistance(string2, maxDistance)
	default:
		panic("unknown DistanceAlgorithm")
	}
}

// DamerauLevenshteinDistance ports the SymSpell optimized Damerau-Levenshtein.
func (e *EditDistance) DamerauLevenshteinDistance(string2 string, maxDistance int) int {
	// Java String.length is UTF-16; LT spelling words are BMP — use runes, not UTF-8 bytes.
	if e.baseString == "" {
		if string2 == "" {
			return 0
		}
		return len([]rune(string2))
	}
	if string2 == "" {
		return len([]rune(e.baseString))
	}
	if maxDistance == 0 {
		if e.baseString == string2 {
			return 0
		}
		return -1
	}

	// work on rune slices for unicode safety while matching Java char indexing for BMP
	base := []rune(e.baseString)
	other := []rune(string2)

	// ensure shorter is string1
	var string1, string2r []rune
	if len(base) > len(other) {
		string1, string2r = other, base
	} else {
		string1, string2r = base, other
	}
	sLen := len(string1)
	tLen := len(string2r)

	// common suffix
	for sLen > 0 && string1[sLen-1] == string2r[tLen-1] {
		sLen--
		tLen--
	}
	start := 0
	if (sLen > 0 && string1[0] == string2r[0]) || sLen == 0 {
		for start < sLen && string1[start] == string2r[start] {
			start++
		}
		sLen -= start
		tLen -= start
		if sLen == 0 {
			return tLen
		}
		string2r = string2r[start : start+tLen]
	}
	lenDiff := tLen - sLen
	if maxDistance < 0 || maxDistance > tLen {
		maxDistance = tLen
	} else if lenDiff > maxDistance {
		return -1
	}

	if tLen > len(e.v0) {
		e.v0 = make([]int, tLen)
		e.v2 = make([]int, tLen)
	} else {
		for i := 0; i < tLen; i++ {
			e.v2[i] = 0
		}
	}
	j := 0
	for ; j < maxDistance; j++ {
		e.v0[j] = j + 1
	}
	for ; j < tLen; j++ {
		e.v0[j] = maxDistance + 1
	}

	jStartOffset := maxDistance - (tLen - sLen)
	haveMax := maxDistance < tLen
	jStart := 0
	jEnd := maxDistance
	sChar := string1[0]
	current := 0
	for i := 0; i < sLen; i++ {
		prevsChar := sChar
		sChar = string1[start+i]
		tChar := string2r[0]
		left := i
		current = left + 1
		nextTransCost := 0
		if i > jStartOffset {
			jStart++
		}
		if jEnd < tLen {
			jEnd++
		}
		for j = jStart; j < jEnd; j++ {
			above := current
			thisTransCost := nextTransCost
			nextTransCost = e.v2[j]
			current = left
			e.v2[j] = current
			left = e.v0[j]
			prevtChar := tChar
			tChar = string2r[j]
			if sChar != tChar {
				if left < current {
					current = left
				}
				if above < current {
					current = above
				}
				current++
				if i != 0 && j != 0 && sChar == prevtChar && prevsChar == tChar {
					thisTransCost++
					if thisTransCost < current {
						current = thisTransCost
					}
				}
			}
			e.v0[j] = current
		}
		if haveMax && e.v0[i+lenDiff] > maxDistance {
			return -1
		}
	}
	if current <= maxDistance {
		return current
	}
	return -1
}
