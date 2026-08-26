package implementation

// DistanceAlgorithm ports EditDistance.DistanceAlgorithm (nested Java enum).
// Java only defines Damerau; other values panic in Compare like IllegalArgumentException.
type DistanceAlgorithm int

const (
	// Damerau is DistanceAlgorithm.Damerau — the only algorithm in upstream.
	Damerau DistanceAlgorithm = iota
)

// EditDistance ports org.languagetool.rules.spelling.symspell.implementation.EditDistance.
//
// Null-string mapping (Go has no null String):
//   - Java ctor empty base → baseString=null: Go sets baseNull=true and baseString="".
//   - Java compare null base + null string2 → 0; null base + non-null → string2.length():
//     Go empty string2 with baseNull → 0; non-empty → javaStringLen(string2).
//   - Java null or empty string2 with non-null base → baseString.length():
//     Go empty string2 → javaStringLen(baseString). Same outcome for null vs empty string2.
// All indexing uses Java String.length/charAt (UTF-16 code units via javaChars).
type EditDistance struct {
	baseString string
	// baseNull is true when Java would hold baseString == null (ctor empty input).
	baseNull  bool
	algorithm DistanceAlgorithm
	v0, v2    []int
}

// NewEditDistance ports EditDistance(String baseString, DistanceAlgorithm algorithm).
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

// Compare ports EditDistance.compare(String string2, int maxDistance).
// Unknown algorithm panics (Java: throw new IllegalArgumentException("unknown DistanceAlgorithm")).
func (e *EditDistance) Compare(string2 string, maxDistance int) int {
	switch e.algorithm {
	case Damerau:
		return e.DamerauLevenshteinDistance(string2, maxDistance)
	default:
		panic("unknown DistanceAlgorithm")
	}
}

// DamerauLevenshteinDistance ports EditDistance.DamerauLevenshteinDistance bug-for-bug
// (optimized windowed Damerau–Levenshtein; UTF-16 length/charAt; -1 if maxDistance exceeded).
func (e *EditDistance) DamerauLevenshteinDistance(string2 string, maxDistance int) int {
	// Java: if (baseString == null) return string2 == null ? 0 : string2.length();
	// Go: empty string2 covers both Java null and empty (both length 0).
	if e.baseNull {
		if string2 == "" {
			return 0
		}
		return javaStringLen(string2)
	}
	// Java: if (string2 == null || string2.isEmpty()) return baseString.length();
	if string2 == "" {
		return javaStringLen(e.baseString)
	}
	// Java: if(maxDistance == 0) return baseString.equals(string2) ? 0 : -1;
	if maxDistance == 0 {
		if e.baseString == string2 {
			return 0
		}
		return -1
	}

	base := javaChars(e.baseString)
	other := javaChars(string2)

	// if strings of different lengths, ensure shorter string is in string1
	var string1, string2c []uint16
	if len(base) > len(other) {
		string1, string2c = other, base
	} else {
		string1, string2c = base, other
	}
	sLen := len(string1)
	tLen := len(string2c)

	// suffix common to both strings can be ignored
	for sLen > 0 && string1[sLen-1] == string2c[tLen-1] {
		sLen--
		tLen--
	}
	start := 0
	// Java: if ((string1.charAt(0) == string2.charAt(0)) || (sLen == 0))
	// After empty early-outs both inputs were non-empty, so charAt(0) is always valid;
	// evaluate sLen==0 first in Go to mirror short-circuit safety when sLen==0.
	if (string1[0] == string2c[0]) || sLen == 0 {
		// prefix common to both strings can be ignored
		for start < sLen && string1[start] == string2c[start] {
			start++
		}
		sLen -= start
		tLen -= start

		// if all of shorter string matches prefix and/or suffix of longer string, then
		// edit distance is just the delete of additional characters present in longer string
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
		// Java: for(int i = 0; i < tLen; i++) v2[i] = 0;
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
		// Java: sChar = string1.charAt(start+i) — string1 is the unsliced shorter original
		sChar = string1[start+i]
		tChar := string2c[0]
		left := i
		current = left + 1
		nextTransCost := 0
		// no need to look beyond window of lower right / upper left diagonals ± maxDistance
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
			// cost of diagonal (substitution)
			current = left
			e.v2[j] = current
			left = e.v0[j] // left now equals current cost (diagonal at next iteration)
			prevtChar := tChar
			tChar = string2c[j]
			if sChar != tChar {
				if left < current {
					current = left // insertion
				}
				if above < current {
					current = above // deletion
				}
				current++
				if i != 0 && j != 0 && sChar == prevtChar && prevsChar == tChar {
					thisTransCost++
					if thisTransCost < current {
						current = thisTransCost // transposition
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
