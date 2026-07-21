package implementation

// DistanceAlgorithm selects the edit distance algorithm.
type DistanceAlgorithm int

const (
	// Damerau is the only algorithm currently supported (Java DistanceAlgorithm.Damerau).
	Damerau DistanceAlgorithm = iota
)

// EditDistance ports org.languagetool.rules.spelling.symspell.implementation.EditDistance.
// All indexing uses Java String length/charAt (UTF-16 code units).
type EditDistance struct {
	baseString string
	// baseNull mirrors Java constructor setting baseString=null when empty input.
	baseNull  bool
	algorithm DistanceAlgorithm
	v0, v2    []int
}

func NewEditDistance(baseString string, algorithm DistanceAlgorithm) *EditDistance {
	e := &EditDistance{baseString: baseString, algorithm: algorithm}
	// Java: if (this.baseString.isEmpty()) { this.baseString = null; return; }
	if baseString == "" {
		e.baseNull = true
		e.baseString = ""
		return e
	}
	if algorithm == Damerau {
		// Java: new int[baseString.length()]
		n := javaStringLen(baseString)
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

// DamerauLevenshteinDistance ports the SymSpell optimized Damerau-Levenshtein
// bug-for-bug with Java (UTF-16 charAt / length).
func (e *EditDistance) DamerauLevenshteinDistance(string2 string, maxDistance int) int {
	// Java: if (baseString == null) return string2 == null ? 0 : string2.length();
	if e.baseNull {
		if string2 == "" {
			// Go has no null strings; empty is the only empty case.
			return 0
		}
		return javaStringLen(string2)
	}
	// Java: if (string2 == null || string2.isEmpty()) return baseString.length();
	if string2 == "" {
		return javaStringLen(e.baseString)
	}
	if maxDistance == 0 {
		if e.baseString == string2 {
			return 0
		}
		return -1
	}

	base := javaChars(e.baseString)
	other := javaChars(string2)

	// ensure shorter is string1
	var string1, string2c []uint16
	if len(base) > len(other) {
		string1, string2c = other, base
	} else {
		string1, string2c = base, other
	}
	sLen := len(string1)
	tLen := len(string2c)

	// common suffix
	for sLen > 0 && string1[sLen-1] == string2c[tLen-1] {
		sLen--
		tLen--
	}
	start := 0
	// Java: if ((string1.charAt(0) == string2.charAt(0)) || (sLen == 0))
	// Note: when sLen==0 after suffix strip, charAt(0) is not evaluated in Java?
	// Actually || short-circuits only if first is true; if sLen==0 and string1 empty after suffix
	// from non-empty, string1 still has content. If original shorter was empty — not here.
	// If sLen==0 because entire string was suffix, string1 may still be non-empty (common case).
	// Edge: if string1 was empty from start — we already returned. Safe when sLen>0 or check length.
	if sLen == 0 || (len(string1) > 0 && string1[0] == string2c[0]) {
		for start < sLen && string1[start] == string2c[start] {
			start++
		}
		sLen -= start
		tLen -= start
		if sLen == 0 {
			return tLen
		}
		// Java: string2 = string2.substring(start, start + tLen)
		string2c = string2c[start : start+tLen]
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
		// Java: sChar = string1.charAt(start+i) — string1 is unsliced original shorter
		sChar = string1[start+i]
		tChar := string2c[0]
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
			tChar = string2c[j]
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
