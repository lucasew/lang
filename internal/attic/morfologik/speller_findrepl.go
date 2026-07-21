package morfologik

import (
	"sort"
	"unicode/utf8"
)

// CandidateData ports morfologik.speller.Speller.CandidateData (word + weighted distance).
type CandidateData struct {
	Word         string
	OrigDistance int // raw edit distance before frequency weight
	Distance     int // distance*FREQ_RANGES + FREQ_RANGES - freq - 1
}

// replPattern ports Speller.Pattern (source chars on the misspelled word side).
type replPattern struct {
	chars        []rune
	startAnchor  bool
	endAnchor    bool
}

// SpellerFSA ports Speller.findReplacementCandidates FSA walk over a Dictionary.
type SpellerFSA struct {
	*SpellerED
	Dict *Dictionary
	// containsSeparators ports Speller.containsSeparators (default true for LT dicts).
	containsSeparators bool
	// candidate buffer (Java char[] candidate)
	candBuf []rune
	// anyToOne: dict 1-char target → patterns matching misspelled multi-char source
	anyToOne map[rune][]replPattern
	// anyToTwo: dict 2-char target string → patterns matching misspelled source
	anyToTwo map[string][]replPattern
}

// NewSpellerFSA builds a Speller-like FSA walker for dict with max edit distance.
func NewSpellerFSA(dict *Dictionary, editDistance int) *SpellerFSA {
	if editDistance < 1 {
		editDistance = 1
	}
	return &SpellerFSA{
		SpellerED:          NewSpellerED(editDistance),
		Dict:               dict,
		containsSeparators: true,
		candBuf:            make([]rune, MaxWordLength),
		anyToOne:           map[rune][]replPattern{},
		anyToTwo:           map[string][]replPattern{},
	}
}

// LoadReplacementPairs ports Speller.createReplacementsMaps for anyToOne/anyToTwo.
// pairs are (from=misspelled side, to=dictionary side) as in fsa.dict.speller.replacement-pairs.
// from may carry ^ / $ anchors.
func (s *SpellerFSA) LoadReplacementPairs(pairs []struct{ From, To string }) {
	if s == nil {
		return
	}
	s.anyToOne = map[rune][]replPattern{}
	s.anyToTwo = map[string][]replPattern{}
	for _, p := range pairs {
		rawKey := p.From
		target := p.To
		if rawKey == "" || target == "" {
			continue
		}
		startA := len(rawKey) > 0 && rawKey[0] == '^'
		endA := len(rawKey) > 0 && rawKey[len(rawKey)-1] == '$'
		stripped := rawKey
		if startA {
			stripped = stripped[1:]
		}
		if endA && len(stripped) > 0 {
			stripped = stripped[:len(stripped)-1]
		}
		pat := replPattern{chars: []rune(stripped), startAnchor: startA, endAnchor: endA}
		tr := []rune(target)
		if len(tr) == 1 {
			s.anyToOne[tr[0]] = append(s.anyToOne[tr[0]], pat)
		} else if len(tr) == 2 {
			key := string(tr)
			s.anyToTwo[key] = append(s.anyToTwo[key], pat)
		}
		// longer targets handled by getAllReplacements outside FSA walk
	}
}

// FindReplacementCandidates ports Speller.findReplacementCandidates(word, false).
// Resets HMatrix each call (Java MorfologikSpeller recreates Speller / HMatrix.reset).
func (s *SpellerFSA) FindReplacementCandidates(word string) []CandidateData {
	if s == nil || s.Dict == nil || s.Dict.FSA == nil || word == "" {
		return nil
	}
	if len(word) == 0 || len(word) >= MaxWordLength {
		return nil
	}
	// evenIfWordInDictionary=false
	if s.Dict.Contains(word) {
		return nil
	}
	s.hMatrix.Reset()
	s.wordProcessed = []rune(word)
	s.wordLen = len(s.wordProcessed)
	if s.wordLen <= s.editDistance {
		s.effectEditDistance = s.wordLen - 1
		if s.effectEditDistance < 0 {
			s.effectEditDistance = 0
		}
	} else {
		s.effectEditDistance = s.editDistance
	}
	s.candidate = s.candBuf
	s.candLen = MaxWordLength

	var candidates []CandidateData
	s.findRepl(&candidates, 0, s.Dict.FSA.RootNode(), nil, 0, 0, -1, "", 0)

	// sort by weighted distance, dedupe words (Java Collections.sort + first occurrence)
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].Distance < candidates[j].Distance
	})
	seen := map[string]struct{}{}
	out := make([]CandidateData, 0, len(candidates))
	for _, c := range candidates {
		if c.Word == "" || c.Word == word {
			continue
		}
		if _, ok := seen[c.Word]; ok {
			continue
		}
		seen[c.Word] = struct{}{}
		out = append(out, c)
	}
	return out
}

// findRepl ports Speller.findRepl (anyToTwo, anyToOne, general).
// Multi-byte labels: accumulate prevBytes until a complete UTF-8 rune decodes.
func (s *SpellerFSA) findRepl(
	candidates *[]CandidateData,
	depth, node int,
	prevBytes []byte,
	wordIndex, candIndex int,
	minLookbackWordIndex int,
	lastAnyToOneSource string,
	lastAnyToOneTarget rune,
) {
	if s == nil || s.Dict == nil || s.Dict.FSA == nil {
		return
	}
	if candIndex >= MaxWordLength-1 || depth > MaxWordLength {
		return
	}
	fsa := s.Dict.FSA
	for arc := fsa.firstArc(node); arc != 0; arc = fsa.nextArc(arc) {
		label := fsa.arcLabel(arc)
		buf := make([]byte, 0, len(prevBytes)+1)
		buf = append(buf, prevBytes...)
		buf = append(buf, label)
		complete, ch := tryDecodeRune(buf)
		if !complete {
			// incomplete multi-byte sequence: descend without advancing depth/wordIndex
			if !fsa.isArcTerminal(arc) {
				s.findRepl(candidates, depth, fsa.endNode(arc), buf, wordIndex, candIndex,
					minLookbackWordIndex, lastAnyToOneSource, lastAnyToOneTarget)
			}
			continue
		}
		// decoded one character into candidate[candIndex]
		s.candBuf[candIndex] = ch

		// --- anyToTwo ---
		if lengthRepl := s.matchAnyToTwo(wordIndex, candIndex, minLookbackWordIndex, lastAnyToOneSource, lastAnyToOneTarget); lengthRepl > 0 {
			if s.isEndOfCandidate(arc, wordIndex) {
				dist := s.hMatrix.Get(depth-1, depth-1)
				if dist <= s.effectEditDistance {
					if extra := absInt(s.wordLen - 1 - (wordIndex + lengthRepl - 2)); extra > 0 {
						dist = dist + extra
					}
					if dist <= s.effectEditDistance {
						w := string(s.candBuf[:candIndex+1])
						*candidates = append(*candidates, s.makeCandidate(w, dist))
					}
				}
			}
			if s.isArcNotTerminal(arc, candIndex) {
				x := s.hMatrix.Get(depth, depth)
				s.hMatrix.Set(depth, depth, s.hMatrix.Get(depth-1, depth-1))
				s.findRepl(candidates, max0(depth), fsa.endNode(arc), nil,
					wordIndex+lengthRepl-1, candIndex+1,
					minLookbackWordIndex, lastAnyToOneSource, lastAnyToOneTarget)
				s.hMatrix.Set(depth, depth, x)
			}
		}

		// --- anyToOne ---
		if lengthRepl := s.matchAnyToOne(wordIndex, candIndex); lengthRepl > 0 {
			if s.isEndOfCandidate(arc, wordIndex) {
				dist := s.hMatrix.Get(depth, depth)
				if dist <= s.effectEditDistance {
					if extra := absInt(s.wordLen - 1 - (wordIndex + lengthRepl - 1)); extra > 0 {
						dist = dist + extra
					}
					if dist <= s.effectEditDistance {
						w := string(s.candBuf[:candIndex+1])
						*candidates = append(*candidates, s.makeCandidate(w, dist))
					}
				}
			}
			if s.isArcNotTerminal(arc, candIndex) {
				newSrc := string(s.wordProcessed[wordIndex : wordIndex+lengthRepl])
				s.findRepl(candidates, depth, fsa.endNode(arc), nil,
					wordIndex+lengthRepl, candIndex+1,
					wordIndex+lengthRepl, newSrc, s.candBuf[candIndex])
			}
		}

		// --- general Oflazer path ---
		if s.Cuted(depth, wordIndex, candIndex) <= s.effectEditDistance {
			if s.isEndOfCandidate(arc, wordIndex) {
				dist := s.Ed(s.wordLen-1-(wordIndex-depth), depth, s.wordLen-1, candIndex)
				if dist <= s.effectEditDistance {
					w := string(s.candBuf[:candIndex+1])
					*candidates = append(*candidates, s.makeCandidate(w, dist))
				}
			}
			if s.isArcNotTerminal(arc, candIndex) {
				s.findRepl(candidates, depth+1, fsa.endNode(arc), nil,
					wordIndex+1, candIndex+1,
					minLookbackWordIndex, lastAnyToOneSource, lastAnyToOneTarget)
			}
		}
	}
}

// matchAnyToOne ports Speller.matchAnyToOne — last candidate char matches multi-char word source.
func (s *SpellerFSA) matchAnyToOne(wordIndex, candIndex int) int {
	if s == nil || len(s.anyToOne) == 0 || candIndex < 0 {
		return 0
	}
	pats, ok := s.anyToOne[s.candBuf[candIndex]]
	if !ok {
		return 0
	}
	for _, p := range pats {
		if p.startAnchor && wordIndex != 0 {
			continue
		}
		i := 0
		for i < len(p.chars) && (wordIndex+i) < s.wordLen && p.chars[i] == s.wordProcessed[wordIndex+i] {
			i++
		}
		if i == len(p.chars) {
			if p.endAnchor && wordIndex+i != s.wordLen {
				continue
			}
			return i
		}
	}
	return 0
}

// matchAnyToTwo ports Speller.matchAnyToTwo — last two candidate chars match word source.
func (s *SpellerFSA) matchAnyToTwo(wordIndex, candIndex, minLookbackWordIndex int, lastAnyToOneSource string, lastAnyToOneTarget rune) int {
	if s == nil || len(s.anyToTwo) == 0 || candIndex < 1 || wordIndex < 1 {
		return 0
	}
	if candIndex >= len(s.candBuf) {
		return 0
	}
	two := string([]rune{s.candBuf[candIndex-1], s.candBuf[candIndex]})
	pats, ok := s.anyToTwo[two]
	if !ok {
		return 0
	}
	for _, p := range pats {
		if p.startAnchor && wordIndex-1 != 0 {
			continue
		}
		// unnecessary replacements when candidate already equals word slice
		if len(p.chars) == 2 && wordIndex < s.wordLen &&
			s.candBuf[candIndex-1] == s.wordProcessed[wordIndex-1] &&
			s.candBuf[candIndex] == s.wordProcessed[wordIndex] {
			return 0
		}
		i := 0
		for i < len(p.chars) && (wordIndex-1+i) < s.wordLen && p.chars[i] == s.wordProcessed[wordIndex-1+i] {
			i++
		}
		if i == len(p.chars) {
			if p.endAnchor && wordIndex-1+i != s.wordLen {
				continue
			}
			// Reject reverse of previous anyToOne at overlapping position
			if wordIndex-1 < minLookbackWordIndex && lastAnyToOneSource != "" &&
				len(p.chars) == 1 && p.chars[0] == lastAnyToOneTarget && two == lastAnyToOneSource {
				continue
			}
			return i
		}
	}
	return 0
}

func max0(d int) int {
	if d < 0 {
		return 0
	}
	return d
}

func (s *SpellerFSA) isArcNotTerminal(arc, candIndex int) bool {
	fsa := s.Dict.FSA
	if fsa.isArcTerminal(arc) {
		return false
	}
	if s.containsSeparators {
		sep := rune(s.Dict.Separator)
		if candIndex >= 0 && s.candBuf[candIndex] == sep {
			return false
		}
	}
	return true
}

func (s *SpellerFSA) isEndOfCandidate(arc, wordIndex int) bool {
	fsa := s.Dict.FSA
	end := fsa.isArcFinal(arc) || s.isBeforeSeparator(arc)
	if !end {
		return false
	}
	return absInt(s.wordLen-1-wordIndex) <= s.effectEditDistance
}

func (s *SpellerFSA) isBeforeSeparator(arc int) bool {
	if !s.containsSeparators || s.Dict == nil {
		return false
	}
	fsa := s.Dict.FSA
	arc1 := fsa.getArc(fsa.endNode(arc), s.Dict.Separator)
	return arc1 != 0 && !fsa.isArcTerminal(arc1)
}

func (s *SpellerFSA) makeCandidate(word string, origDist int) CandidateData {
	freq := 0
	if s.Dict != nil {
		freq = s.Dict.GetFrequency(word)
		if freq < 0 {
			freq = 0
		}
	}
	// CandidateData: distance = orig*FREQ_RANGES + FREQ_RANGES - freq - 1
	dist := origDist*FreqRanges + FreqRanges - freq - 1
	return CandidateData{Word: word, OrigDistance: origDist, Distance: dist}
}

// tryDecodeRune attempts UTF-8 decode of buf. complete=false if need more bytes.
func tryDecodeRune(buf []byte) (complete bool, ch rune) {
	if len(buf) == 0 {
		return false, 0
	}
	r, size := utf8.DecodeRune(buf)
	if r == utf8.RuneError && size == 1 {
		// could be incomplete multi-byte start
		if !utf8.FullRune(buf) {
			return false, 0
		}
		// invalid single byte — treat as that byte
		return true, rune(buf[0])
	}
	if size < len(buf) {
		// extra bytes shouldn't happen when we grow one at a time
		return true, r
	}
	return true, r
}

// WeightedFindReplacements converts CandidateData to the WeightedEditSuggestions shape.
func (s *SpellerFSA) WeightedFindReplacements(word string, maxResults int) []struct {
	Word   string
	Weight int
} {
	cds := s.FindReplacementCandidates(word)
	if len(cds) == 0 {
		return nil
	}
	if maxResults > 0 && len(cds) > maxResults {
		cds = cds[:maxResults]
	}
	out := make([]struct {
		Word   string
		Weight int
	}, len(cds))
	for i, c := range cds {
		out[i].Word = c.Word
		out[i].Weight = c.Distance
	}
	return out
}
