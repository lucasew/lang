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
	// compactMask = (0xffffffff >> (3 + defaultCompactLevel)) << 2
	mask := int((uint32(0xffffffff) >> (3 + defaultCompactLevel)) << 2)
	return &SymSpell{
		initialCapacity:           initialCapacity,
		maxDictionaryEditDistance: maxDictionaryEditDistance,
		prefixLength:              prefixLength,
		countThreshold:            countThreshold,
		compactMask:               mask,
		distanceAlgorithm:         Damerau,
		deletes:                   map[int][]string{},
		words:                     make(map[string]int64, initialCapacity),
		belowThresholdWords:       map[string]int64{},
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

// CreateDictionaryEntry adds/updates a dictionary word. Returns true if a new
// above-threshold word was added (first time edits are generated).
func (s *SymSpell) CreateDictionaryEntry(key string, count int64, staging *SuggestionStage) bool {
	if s == nil {
		return false
	}
	if count <= 0 {
		if s.countThreshold > 0 {
			return false
		}
		count = 0
	}

	if s.countThreshold > 1 {
		if prev, ok := s.belowThresholdWords[key]; ok {
			if count > 0 && prev > (1<<63-1)-count {
				count = 1<<63 - 1
			} else {
				count = prev + count
			}
			if count >= s.countThreshold {
				delete(s.belowThresholdWords, key)
			} else {
				s.belowThresholdWords[key] = count
				return false
			}
		} else if prev, ok := s.words[key]; ok {
			if count > 0 && prev > (1<<63-1)-count {
				count = 1<<63 - 1
			} else {
				count = prev + count
			}
			s.words[key] = count
			return false
		} else if count < s.countThreshold {
			s.belowThresholdWords[key] = count
			return false
		}
	} else {
		if prev, ok := s.words[key]; ok {
			if count > 0 && prev > (1<<63-1)-count {
				count = 1<<63 - 1
			} else {
				count = prev + count
			}
			s.words[key] = count
			return false
		} else if count < s.countThreshold {
			s.belowThresholdWords[key] = count
			return false
		}
	}

	s.words[key] = count
	if javaStringLen(key) > s.maxLength {
		s.maxLength = javaStringLen(key)
	}

	edits := s.editsPrefix(key)
	if staging != nil {
		for del := range edits {
			staging.Add(s.stringHash(del), key)
		}
	} else {
		for del := range edits {
			h := s.stringHash(del)
			s.deletes[h] = append(s.deletes[h], key)
		}
	}
	return true
}

// CommitStaging flushes a SuggestionStage into permanent deletes.
func (s *SymSpell) CommitStaging(staging *SuggestionStage) {
	if s == nil || staging == nil {
		return
	}
	if s.deletes == nil {
		s.deletes = map[int][]string{}
	}
	staging.CommitTo(s.deletes)
}

// Lookup finds spelling suggestions for input.
func (s *SymSpell) Lookup(input string, verbosity Verbosity) []SuggestItem {
	return s.LookupMax(input, verbosity, s.maxDictionaryEditDistance)
}

// LookupMax finds suggestions with an explicit max edit distance.
func (s *SymSpell) LookupMax(input string, verbosity Verbosity, maxEditDistance int) []SuggestItem {
	if s == nil {
		return nil
	}
	if maxEditDistance > s.maxDictionaryEditDistance {
		panic("Dist too big: " + itoa(maxEditDistance))
	}

	var suggestions []SuggestItem
	inputLen := javaStringLen(input)
	if inputLen-maxEditDistance > s.maxLength {
		return suggestions
	}

	consideredDeletes := map[string]struct{}{}
	consideredSuggestions := map[string]struct{}{}

	if c, ok := s.words[input]; ok {
		suggestions = append(suggestions, NewSuggestItem(input, 0, c))
		if verbosity != VerbosityAll {
			return suggestions
		}
	}
	consideredSuggestions[input] = struct{}{}

	maxEditDistance2 := maxEditDistance
	candidates := []string{}
	inputPrefixLen := inputLen
	if inputPrefixLen > s.prefixLength {
		inputPrefixLen = s.prefixLength
		candidates = append(candidates, javaSubstring(input, 0, inputPrefixLen))
	} else {
		candidates = append(candidates, input)
	}

	distanceComparer := NewEditDistance(input, s.distanceAlgorithm)
	candidatePointer := 0
	for candidatePointer < len(candidates) {
		candidate := candidates[candidatePointer]
		candidatePointer++
		candidateLen := javaStringLen(candidate)
		lengthDiff := inputPrefixLen - candidateLen

		if lengthDiff > maxEditDistance2 {
			if verbosity == VerbosityAll {
				continue
			}
			break
		}

		if dictSuggestions, ok := s.deletes[s.stringHash(candidate)]; ok {
			for _, suggestion := range dictSuggestions {
				if suggestion == input {
					continue
				}
				suggestionLen := javaStringLen(suggestion)
				if abs(suggestionLen-inputLen) > maxEditDistance2 ||
					suggestionLen < candidateLen ||
					(suggestionLen == candidateLen && suggestion != candidate) {
					continue
				}
				suggPrefixLen := suggestionLen
				if suggPrefixLen > s.prefixLength {
					suggPrefixLen = s.prefixLength
				}
				if suggPrefixLen > inputPrefixLen && (suggPrefixLen-candidateLen) > maxEditDistance2 {
					continue
				}

				var distance int
				if candidateLen == 0 {
					distance = max(inputLen, suggestionLen)
					if distance > maxEditDistance2 {
						continue
					}
					if _, seen := consideredSuggestions[suggestion]; seen {
						continue
					}
					consideredSuggestions[suggestion] = struct{}{}
				} else if suggestionLen == 1 {
					// Java: input.indexOf(suggestion.charAt(0))
					ch := javaChars(suggestion)[0]
					if javaIndexOfChar(input, ch) < 0 {
						distance = inputLen
					} else {
						distance = inputLen - 1
					}
					if distance > maxEditDistance2 {
						continue
					}
					if _, seen := consideredSuggestions[suggestion]; seen {
						continue
					}
					consideredSuggestions[suggestion] = struct{}{}
				} else {
					if verbosity != VerbosityAll && !s.deleteInSuggestionPrefix(candidate, candidateLen, suggestion, suggestionLen) {
						continue
					}
					if _, seen := consideredSuggestions[suggestion]; seen {
						continue
					}
					consideredSuggestions[suggestion] = struct{}{}
					distance = distanceComparer.Compare(suggestion, maxEditDistance2)
					if distance < 0 {
						continue
					}
				}

				if distance <= maxEditDistance2 {
					suggestionCount := s.words[suggestion]
					si := NewSuggestItem(suggestion, distance, suggestionCount)
					if len(suggestions) > 0 {
						switch verbosity {
						case VerbosityClosest:
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

		if lengthDiff < maxEditDistance && candidateLen <= s.prefixLength {
			if verbosity != VerbosityAll && lengthDiff >= maxEditDistance2 {
				continue
			}
			for i := 0; i < candidateLen; i++ {
				// Java: StringBuilder.deleteCharAt(i) — UTF-16 unit
				del := javaDeleteCharAt(candidate, i)
				if _, ok := consideredDeletes[del]; !ok {
					consideredDeletes[del] = struct{}{}
					candidates = append(candidates, del)
				}
			}
		}
	}

	if len(suggestions) > 1 {
		sort.Slice(suggestions, func(i, j int) bool {
			return suggestions[i].Less(suggestions[j])
		})
	}
	return suggestions
}

func (s *SymSpell) deleteInSuggestionPrefix(delete string, deleteLen int, suggestion string, suggestionLen int) bool {
	if deleteLen == 0 {
		return true
	}
	if s.prefixLength < suggestionLen {
		suggestionLen = s.prefixLength
	}
	delU := javaChars(delete)
	sugU := javaChars(suggestion)
	if suggestionLen > len(sugU) {
		suggestionLen = len(sugU)
	}
	j := 0
	for i := 0; i < deleteLen && i < len(delU); i++ {
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

func (s *SymSpell) edits(word string, editDistance int, deleteWords map[string]struct{}) {
	editDistance++
	// Java: if (word.length() > 1) for i in 0..length-1 deleteCharAt(i)
	wLen := javaStringLen(word)
	if wLen > 1 {
		for i := 0; i < wLen; i++ {
			del := javaDeleteCharAt(word, i)
			if _, ok := deleteWords[del]; !ok {
				deleteWords[del] = struct{}{}
				if editDistance < s.maxDictionaryEditDistance {
					s.edits(del, editDistance, deleteWords)
				}
			}
		}
	}
}

func (s *SymSpell) editsPrefix(key string) map[string]struct{} {
	hashSet := map[string]struct{}{}
	if javaStringLen(key) <= s.maxDictionaryEditDistance {
		hashSet[""] = struct{}{}
	}
	if javaStringLen(key) > s.prefixLength {
		key = javaSubstring(key, 0, s.prefixLength)
	}
	hashSet[key] = struct{}{}
	s.edits(key, 0, hashSet)
	return hashSet
}

func (s *SymSpell) stringHash(str string) int {
	// Java getStringHash: length and charAt are UTF-16
	u := javaChars(str)
	lenMask := len(u)
	if lenMask > 3 {
		lenMask = 3
	}
	var hash uint32 = 2166136261
	for i := 0; i < len(u); i++ {
		hash ^= uint32(u[i])
		hash *= 16777619
	}
	hash &= uint32(s.compactMask)
	hash |= uint32(lenMask)
	return int(hash)
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
