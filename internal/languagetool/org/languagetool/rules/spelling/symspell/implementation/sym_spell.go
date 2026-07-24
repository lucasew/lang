package implementation

import (
	"sort"
)

// Verbosity controls Lookup result quantity/closeness.
type Verbosity int

const (
	VerbosityTop Verbosity = iota
	VerbosityClosest
	VerbosityAll
)

const (
	defaultMaxEditDistance = 2
	defaultPrefixLength    = 7
	defaultCountThreshold  = 1
	defaultInitialCapacity = 16
	defaultCompactLevel    = 5
)

// SymSpell ports org.languagetool.rules.spelling.symspell.implementation.SymSpell.
// Core surface: dictionary build + single-word lookup.
type SymSpell struct {
	initialCapacity           int
	maxDictionaryEditDistance int
	prefixLength              int
	countThreshold            int64
	compactMask               int
	distanceAlgorithm         DistanceAlgorithm
	maxLength                 int
	deletes                   map[int][]string
	words                     map[string]int64
	belowThresholdWords       map[string]int64
}

func NewSymSpell(initialCapacity, maxDictionaryEditDistance, prefixLength int, countThreshold int64) *SymSpell {
	if initialCapacity < 0 {
		initialCapacity = defaultInitialCapacity
	}
	if maxDictionaryEditDistance < 0 {
		maxDictionaryEditDistance = defaultMaxEditDistance
	}
	if prefixLength < 1 || prefixLength <= maxDictionaryEditDistance {
		prefixLength = defaultPrefixLength
	}
	if countThreshold < 0 {
		countThreshold = defaultCountThreshold
	}
	// Java: (0xffffffff >> (3 + defaultCompactLevel)) << 2
	// 0xffffffff is signed int -1; >> is arithmetic → result -4 (not uint32 logical shift).
	mask := int((int32(-1) >> (3 + defaultCompactLevel)) << 2)
	return &SymSpell{
		initialCapacity:           initialCapacity,
		maxDictionaryEditDistance: maxDictionaryEditDistance,
		prefixLength:              prefixLength,
		countThreshold:            countThreshold,
		compactMask:               mask,
		distanceAlgorithm:         Damerau,
		// Java leaves deletes null until first create/load/commit.
		deletes:             nil,
		words:               make(map[string]int64, initialCapacity),
		belowThresholdWords: map[string]int64{},
	}
}

func DefaultSymSpell() *SymSpell {
	return NewSymSpell(defaultInitialCapacity, defaultMaxEditDistance, defaultPrefixLength, defaultCountThreshold)
}

func (s *SymSpell) WordCount() int {
	if s == nil {
		return 0
	}
	return len(s.words)
}

func (s *SymSpell) MaxDictionaryEditDistance() int {
	if s == nil {
		return 0
	}
	return s.maxDictionaryEditDistance
}

// CreateDictionaryEntry ports createDictionaryEntry(key, count, staging).
// Returns true if the word was added as a new correctly spelled word,
// or false if below threshold / update of existing above-threshold word.
func (s *SymSpell) CreateDictionaryEntry(key string, count int64, staging *SuggestionStage) bool {
	if s == nil {
		return false
	}
	if count <= 0 {
		if s.countThreshold > 0 {
			return false // no point doing anything if count is zero, as it can't change anything
		}
		count = 0
	}

	// look first in below threshold words, update count, and allow promotion
	// threshold must be >1 for there to be the possibility of low threshold words
	if s.countThreshold > 1 {
		if countPrevious, ok := s.belowThresholdWords[key]; ok {
			// count = (Long.MAX_VALUE - countPrevious > count) ? countPrevious + count : Long.MAX_VALUE
			if int64(^uint64(0)>>1)-countPrevious > count {
				count = countPrevious + count
			} else {
				count = int64(^uint64(0) >> 1) // Long.MAX_VALUE
			}
			if count >= s.countThreshold {
				delete(s.belowThresholdWords, key)
			} else {
				s.belowThresholdWords[key] = count
				return false
			}
		} else if countPrevious, ok := s.words[key]; ok {
			if int64(^uint64(0)>>1)-countPrevious > count {
				count = countPrevious + count
			} else {
				count = int64(^uint64(0) >> 1)
			}
			s.words[key] = count
			return false
		} else if count < s.countThreshold {
			s.belowThresholdWords[key] = count
			return false
		}
	} else if countPrevious, ok := s.words[key]; ok {
		// Java: else if (words.containsKey(key)) when countThreshold <= 1
		// (first branch short-circuits on countThreshold > 1)
		if int64(^uint64(0)>>1)-countPrevious > count {
			count = countPrevious + count
		} else {
			count = int64(^uint64(0) >> 1)
		}
		s.words[key] = count
		return false
	} else if count < s.countThreshold {
		s.belowThresholdWords[key] = count
		return false
	}

	// what we have at this point is a new, above threshold word
	s.words[key] = count

	//edits/suggestions are created only once, no matter how often word occurs
	if javaStringLen(key) > s.maxLength {
		s.maxLength = javaStringLen(key)
	}

	edits := s.editsPrefix(key)
	// if not staging suggestions, put directly into main data structure
	if staging != nil {
		for del := range edits {
			staging.Add(s.getStringHash(del), key)
		}
	} else {
		if s.deletes == nil {
			s.deletes = make(map[int][]string, s.initialCapacity)
		}
		for del := range edits {
			deleteHash := s.getStringHash(del)
			suggestions, ok := s.deletes[deleteHash]
			if ok {
				// Java: Arrays.copyOf(suggestions, suggestions.length + 1)
				newSuggestions := make([]string, len(suggestions)+1)
				copy(newSuggestions, suggestions)
				s.deletes[deleteHash] = newSuggestions
				suggestions = newSuggestions
			} else {
				suggestions = make([]string, 1)
				s.deletes[deleteHash] = suggestions
			}
			suggestions[len(suggestions)-1] = key
		}
	}
	return true
}

// CommitStaging flushes a SuggestionStage into permanent deletes.
// Ports commitStaged: if deletes == null, new HashMap<>(staging.deletes.size()).
func (s *SymSpell) CommitStaging(staging *SuggestionStage) {
	if s == nil || staging == nil {
		return
	}
	if s.deletes == nil {
		s.deletes = make(map[int][]string, len(staging.Deletes))
	}
	staging.CommitTo(s.deletes)
}

// Lookup finds spelling suggestions for input using maxDictionaryEditDistance.
func (s *SymSpell) Lookup(input string, verbosity Verbosity) []SuggestItem {
	return s.LookupMax(input, verbosity, s.maxDictionaryEditDistance)
}

// LookupMax ports lookup(input, verbosity, maxEditDistance).
func (s *SymSpell) LookupMax(input string, verbosity Verbosity, maxEditDistance int) []SuggestItem {
	if s == nil {
		return nil
	}
	// maxEditDistance used in lookup can't be bigger than the maxDictionaryEditDistance
	if maxEditDistance > s.maxDictionaryEditDistance {
		// Java: throw new IllegalArgumentException("Dist to big: " + maxEditDistance);
		panic("Dist to big: " + itoa(maxEditDistance))
	}

	suggestions := make([]SuggestItem, 0)
	inputLen := javaStringLen(input)

	// early exit - word is too big to possibly match any words
	if inputLen-maxEditDistance > s.maxLength {
		return suggestions
	}

	// deletes we've considered already
	consideredDeletes := map[string]struct{}{}
	// suggestions we've considered already
	consideredSuggestions := map[string]struct{}{}

	// quick look for exact match
	if suggestionCount, ok := s.words[input]; ok {
		suggestions = append(suggestions, NewSuggestItem(input, 0, suggestionCount))
		// early exit - return exact match, unless caller wants all matches
		if verbosity != VerbosityAll {
			return suggestions
		}
	}
	consideredSuggestions[input] = struct{}{} // input considered in above.

	maxEditDistance2 := maxEditDistance
	candidatePointer := 0
	candidates := make([]string, 0)

	// add original prefix
	inputPrefixLen := inputLen
	if inputPrefixLen > s.prefixLength {
		inputPrefixLen = s.prefixLength
		candidates = append(candidates, javaSubstring(input, 0, inputPrefixLen))
	} else {
		candidates = append(candidates, input)
	}

	distanceComparer := NewEditDistance(input, s.distanceAlgorithm)
	for candidatePointer < len(candidates) {
		candidate := candidates[candidatePointer]
		candidatePointer++
		candidateLen := javaStringLen(candidate)
		lengthDiff := inputPrefixLen - candidateLen

		// early termination if distance higher than suggestion distance
		if lengthDiff > maxEditDistance2 {
			// skip to next candidate if Verbosity.All, look no further if Top or Closest
			if verbosity == VerbosityAll {
				continue
			}
			break
		}

		// read candidate entry from dictionary (nil deletes map: Go read = miss, like no entry)
		if dictSuggestions, ok := s.deletes[s.getStringHash(candidate)]; ok {
			for _, suggestion := range dictSuggestions {
				if suggestion == input {
					continue
				}
				suggestionLen := javaStringLen(suggestion)

				if abs(suggestionLen-inputLen) > maxEditDistance2 || // input/suggestion diff > allowed
					suggestionLen < candidateLen || // sugg must be for a different delete String
					(suggestionLen == candidateLen && suggestion != candidate) { // hash collision
					continue
				}

				suggPrefixLen := suggestionLen
				if suggPrefixLen > s.prefixLength {
					suggPrefixLen = s.prefixLength
				}
				if suggPrefixLen > inputPrefixLen && (suggPrefixLen-candidateLen) > maxEditDistance2 {
					continue
				}

				// True Damerau-Levenshtein Edit Distance: adjust distance, if both distances > 0
				var distance int
				min := 0
				if candidateLen == 0 {
					// suggestions which have no common chars with input
					distance = max(inputLen, suggestionLen)
					if distance > maxEditDistance2 || !setAdd(consideredSuggestions, suggestion) {
						continue
					}
				} else if suggestionLen == 1 {
					// Java: input.indexOf(suggestion.charAt(0))
					ch := javaChars(suggestion)[0]
					if javaIndexOfChar(input, ch) < 0 {
						distance = inputLen
					} else {
						distance = inputLen - 1
					}
					if distance > maxEditDistance2 || !setAdd(consideredSuggestions, suggestion) {
						continue
					}
				} else {
					// number of edits in prefix == maxeditdistance && no identical suffix
					// then editdistance > maxEditDistance and no need for Levenshtein calculation
					// Java: if ((A && B) || C) continue; else { distance calc }
					// A = (prefixLength - maxEditDistance == candidateLen)
					// B assigns min = min(inputLen,suggestionLen) - prefixLength; then min > 1 && !suffix equal
					// C = (min > 0) && char-mismatch transpose check (min stays 0 if A was false)
					skipDistance := false
					if s.prefixLength-maxEditDistance == candidateLen {
						min = minInt(inputLen, suggestionLen) - s.prefixLength
						if min > 1 && javaSubstring(input, inputLen+1-min, inputLen) != javaSubstring(suggestion, suggestionLen+1-min, suggestionLen) {
							skipDistance = true
						}
					}
					if !skipDistance && min > 0 &&
						javaCharAt(input, inputLen-min) != javaCharAt(suggestion, suggestionLen-min) &&
						(javaCharAt(input, inputLen-min-1) != javaCharAt(suggestion, suggestionLen-min) ||
							javaCharAt(input, inputLen-min) != javaCharAt(suggestion, suggestionLen-min-1)) {
						skipDistance = true
					}
					if skipDistance {
						continue
					}
					// deleteInSuggestionPrefix is somewhat expensive, and only pays off when verbosity is Top or Closest.
					if (verbosity != VerbosityAll && !s.deleteInSuggestionPrefix(candidate, candidateLen, suggestion, suggestionLen)) ||
						!setAdd(consideredSuggestions, suggestion) {
						continue
					}
					distance = distanceComparer.Compare(suggestion, maxEditDistance2)
					if distance < 0 {
						continue
					}
				}

				// save some time
				// do not process higher distances than those already found, if verbosity<All
				if distance <= maxEditDistance2 {
					suggestionCount := s.words[suggestion]
					si := NewSuggestItem(suggestion, distance, suggestionCount)
					if len(suggestions) > 0 {
						switch verbosity {
						case VerbosityClosest:
							// we will calculate DamLev distance only to the smallest found distance so far
							if distance < maxEditDistance2 {
								suggestions = suggestions[:0]
							}
						case VerbosityTop:
							if distance < maxEditDistance2 || suggestionCount > suggestions[0].Count {
								maxEditDistance2 = distance
								suggestions[0] = si
							}
							continue
						}
					}
					if verbosity != VerbosityAll {
						maxEditDistance2 = distance
					}
					suggestions = append(suggestions, si)
				}
			}
		}

		// add edits: derive deletes from candidate and add to candidates list
		if lengthDiff < maxEditDistance && candidateLen <= s.prefixLength {
			// do not create edits with edit distance smaller than suggestions already found
			if verbosity != VerbosityAll && lengthDiff >= maxEditDistance2 {
				continue
			}
			for i := 0; i < candidateLen; i++ {
				// Java: StringBuilder.deleteCharAt(i) — UTF-16 unit
				del := javaDeleteCharAt(candidate, i)
				if setAdd(consideredDeletes, del) {
					candidates = append(candidates, del)
				}
			}
		}
	}

	// sort by ascending edit distance, then by descending word frequency
	// Java Collections.sort is stable
	if len(suggestions) > 1 {
		sort.SliceStable(suggestions, func(i, j int) bool {
			return suggestions[i].Less(suggestions[j])
		})
	}
	return suggestions
}

// setAdd ports HashSet.add: inserts and returns true if not already present.
func setAdd(set map[string]struct{}, v string) bool {
	if _, ok := set[v]; ok {
		return false
	}
	set[v] = struct{}{}
	return true
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// deleteInSuggestionPrefix ports private deleteInSuggestionPrefix.
// Every delete string must be a valid prefix of a dictionary suggestion up to prefixLength.
func (s *SymSpell) deleteInSuggestionPrefix(delete string, deleteLen int, suggestion string, suggestionLen int) bool {
	if deleteLen == 0 {
		return true
	}
	if s.prefixLength < suggestionLen {
		suggestionLen = s.prefixLength
	}
	delU := javaChars(delete)
	sugU := javaChars(suggestion)
	j := 0
	// Java: for (int i = 0; i < deleteLen; i++) with charAt — assumes lengths match
	for i := 0; i < deleteLen; i++ {
		delChar := delU[i]
		for j < suggestionLen && delChar != sugU[j] {
			j++
		}
		if j == suggestionLen {
			return false
		}
	}
	return true
}

// edits ports private edits(word, editDistance, deleteWords) — recursive deletes.
func (s *SymSpell) edits(word string, editDistance int, deleteWords map[string]struct{}) map[string]struct{} {
	editDistance++
	// Java: if (word.length() > 1) for i in 0..length-1 deleteCharAt(i)
	wLen := javaStringLen(word)
	if wLen > 1 {
		for i := 0; i < wLen; i++ {
			del := javaDeleteCharAt(word, i)
			if _, ok := deleteWords[del]; !ok {
				deleteWords[del] = struct{}{}
				// recursion, if maximum edit distance not yet reached
				if editDistance < s.maxDictionaryEditDistance {
					s.edits(del, editDistance, deleteWords)
				}
			}
		}
	}
	return deleteWords
}

// editsPrefix ports private editsPrefix(key).
func (s *SymSpell) editsPrefix(key string) map[string]struct{} {
	hashSet := map[string]struct{}{}
	if javaStringLen(key) <= s.maxDictionaryEditDistance {
		hashSet[""] = struct{}{}
	}
	if javaStringLen(key) > s.prefixLength {
		key = javaSubstring(key, 0, s.prefixLength)
	}
	hashSet[key] = struct{}{}
	return s.edits(key, 0, hashSet)
}

// getStringHash ports private getStringHash(String s).
// Java uses long FNV-like multiply (64-bit), then &= compactMask (int promoted with sign-extend),
// then |= lenMask, cast to int. compactMask is -4 under defaultCompactLevel=5.
func (s *SymSpell) getStringHash(str string) int {
	u := javaChars(str)
	lenMask := len(u)
	if lenMask > 3 {
		lenMask = 3
	}
	var hash int64 = 2166136261
	for i := 0; i < len(u); i++ {
		hash ^= int64(u[i])
		hash *= 16777619
	}
	// Java: hash &= this.compactMask; int → long sign-extends (e.g. -4 → all-bits-but-low-2)
	hash &= int64(int32(s.compactMask))
	hash |= int64(lenMask)
	return int(int32(hash))
}

// stringHash is an alias kept for any older call sites; same as getStringHash.
func (s *SymSpell) stringHash(str string) int {
	return s.getStringHash(str)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
