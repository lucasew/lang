package multitoken

import (
	"bufio"
	"io"
	"strings"
	"sync"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const maxLengthDiff = 3

// WeightedSuggestion ports morfologik.WeightedSuggestion for multitoken merge path.
type WeightedSuggestion struct {
	Word   string
	Weight int
}

// MultitokenSpeller ports org.languagetool.rules.spelling.multitoken.MultitokenSpeller
// as a dictionary-backed multiword suggestion engine (no full Language/speller stack).
type MultitokenSpeller struct {
	// first char of normalized key → normalizedKey → original lines
	byFirstChar map[rune]map[string][]string
	// normalized key without spaces → original lines
	noSpaces map[string][]string
	// PrepareLine optional language hook (default: identity single line).
	PrepareLine func(line string) []string
	// IsException optional language hook (Java MultitokenSpeller.isException).
	// When true for (original, candidate), stopSearching returns true (no suggestion).
	IsException func(original, candidate string) bool
	// GetAdditionalSuggestions ports getAdditionalSuggestions (e.g. Catalan Morfologik).
	// If any additional word equals originalWord, Java returns empty list (exact hit).
	GetAdditionalSuggestions func(originalWord string) []WeightedSuggestion
	// IsMisspelledToken ports SpellingCheckRule.isMisspelled for discardRunOnWords.
	// Nil → discardRunOnWords returns false (cannot detect run-ons without a speller).
	IsMisspelledToken func(token string) bool
	mu                sync.RWMutex
	cache             map[string][]string
}

func NewMultitokenSpeller() *MultitokenSpeller {
	return &MultitokenSpeller{
		byFirstChar: map[rune]map[string][]string{},
		noSpaces:    map[string][]string{},
		cache:       map[string][]string{},
		PrepareLine: func(line string) []string { return []string{line} },
	}
}

// LoadWords loads multiword dictionary lines from r (skip #/empty, strip trailing comments).
// Single-token lines are ignored (same as Java).
func (m *MultitokenSpeller) LoadWords(r io.Reader) error {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		lineOriginal := sc.Text()
		if lineOriginal == "" || lineOriginal[0] == '#' {
			continue
		}
		lineOriginal = strings.TrimSpace(lineOriginal)
		if i := strings.Index(lineOriginal, "#"); i >= 0 {
			lineOriginal = strings.TrimSpace(lineOriginal[:i])
		}
		if lineOriginal == "" {
			continue
		}
		prep := m.PrepareLine
		if prep == nil {
			prep = func(line string) []string { return []string{line} }
		}
		for _, line := range prep(lineOriginal) {
			if line == "" {
				continue
			}
			normalizedKey := getNormalizeKey(line)
			if !strings.Contains(normalizedKey, " ") {
				continue
			}
			first := []rune(normalizedKey)[0]
			m.mu.Lock()
			if m.byFirstChar[first] == nil {
				m.byFirstChar[first] = map[string][]string{}
			}
			addToMap(m.byFirstChar[first], normalizedKey, line)
			addToMap(m.noSpaces, strings.ReplaceAll(normalizedKey, " ", ""), line)
			m.mu.Unlock()
		}
	}
	return sc.Err()
}

// GetSuggestions returns multiword spelling suggestions for originalWord.
func (m *MultitokenSpeller) GetSuggestions(originalWord string) []string {
	return m.GetSuggestionsOpts(originalWord, false)
}

func (m *MultitokenSpeller) GetSuggestionsOpts(originalWord string, areTokensAcceptedBySpeller bool) []string {
	originalWord = collapseWhitespace(originalWord)
	m.mu.RLock()
	if cached, ok := m.cache[originalWord]; ok {
		m.mu.RUnlock()
		return append([]string(nil), cached...)
	}
	m.mu.RUnlock()

	results := m.computeSuggestions(originalWord, areTokensAcceptedBySpeller)
	m.mu.Lock()
	m.cache[originalWord] = results
	m.mu.Unlock()
	return append([]string(nil), results...)
}

type weighted struct {
	word   string
	weight int
}

func (m *MultitokenSpeller) computeSuggestions(originalWord string, areTokensAcceptedBySpeller bool) []string {
	// Java: word = originalWord.replace("- ", "-").replace(" -", "-");
	word := strings.ReplaceAll(strings.ReplaceAll(originalWord, "- ", "-"), " -", "-")
	if word == "" {
		return nil
	}
	// Java discardRunOnWords(word)
	if m.discardRunOnWords(word) {
		return nil
	}
	normalizedWord := getNormalizeKey(word)
	if normalizedWord == "" {
		return nil
	}
	var weightedCandidates []weighted

	normalizedNoSpaces := strings.ReplaceAll(normalizedWord, " ", "")
	m.mu.RLock()
	if cands, ok := m.noSpaces[normalizedNoSpaces]; ok {
		if m.stopSearching(cands, originalWord) {
			m.mu.RUnlock()
			return nil
		}
		for _, c := range cands {
			weightedCandidates = append(weightedCandidates, weighted{c, 0})
		}
	}
	// Java: Character firstChar = normalizedWord.charAt(0); String UTF-16
	// Use first rune for non-BMP safety (same for BMP multiword lists).
	first := []rune(normalizedWord)[0]
	byChar := m.byFirstChar[first]
	if len(weightedCandidates) == 0 && byChar != nil {
		for normalizedCandidate, candidates := range byChar {
			if m.stopSearching(candidates, originalWord) {
				m.mu.RUnlock()
				return nil
			}
			// Java: Math.abs(normalizedCandidate.length() - word.length()) — UTF-16 lengths
			if abs(utf16Len(normalizedCandidate)-utf16Len(word)) > maxLengthDiff {
				continue
			}
			candidateParts := splitBySpace(normalizedCandidate)
			wordParts := splitBySpace(normalizedWord)
			distances := distancesPerWord(candidateParts, wordParts, normalizedCandidate, normalizedWord)
			totalDistance := 0
			for _, d := range distances {
				totalDistance += d
			}
			if totalDistance < 1 {
				for _, c := range candidates {
					weightedCandidates = append(weightedCandidates, weighted{c, 0})
				}
				// Java: "continue" allows several candidates with different casing
				if len(weightedCandidates) == 2 {
					break
				}
				continue
			}
			// for very short candidates, allow only distance=0
			if utf16Len(normalizedCandidate) < 7 {
				continue
			}
			exceedsMaxDistancePerToken := false
			for i := 0; i < len(distances); i++ {
				// usually 2, but 1 for short words
				maxDist := 1
				if i < len(wordParts) && i < len(candidateParts) &&
					utf16Len(wordParts[i]) > 5 && utf16Len(candidateParts[i]) > 4 {
					maxDist = 2
				}
				if distances[i] > maxDist {
					exceedsMaxDistancePerToken = true
					break
				}
			}
			if exceedsMaxDistancePerToken {
				continue
			}
			if totalDistance <= maxEditDistance(normalizedCandidate, normalizedWord) {
				for _, c := range candidates {
					weightedCandidates = append(weightedCandidates, weighted{c, totalDistance})
				}
			}
		}
	}
	m.mu.RUnlock()

	// Java: for (WeightedSuggestion additionalSuggestion : getAdditionalSuggestions(word))
	if m.GetAdditionalSuggestions != nil {
		for _, add := range m.GetAdditionalSuggestions(word) {
			if add.Word == "" {
				continue
			}
			if add.Word == originalWord {
				// Java: return Collections.emptyList() when additional equals original
				return nil
			}
			weightedCandidates = append(weightedCandidates, weighted{add.Word, add.Weight})
		}
	}

	if len(weightedCandidates) == 0 {
		return nil
	}
	// sort by weight (stable enough for twin tests)
	for i := 0; i < len(weightedCandidates); i++ {
		for j := i + 1; j < len(weightedCandidates); j++ {
			if weightedCandidates[j].weight < weightedCandidates[i].weight {
				weightedCandidates[i], weightedCandidates[j] = weightedCandidates[j], weightedCandidates[i]
			}
		}
	}
	weightFirst := weightedCandidates[0].weight
	if areTokensAcceptedBySpeller && strings.ToUpper(weightedCandidates[0].word) == originalWord {
		return nil
	}
	if areTokensAcceptedBySpeller && weightFirst > 1 {
		return nil
	}
	var results []string
	seen := map[string]struct{}{}
	for _, w := range weightedCandidates {
		if w.weight-weightFirst < 1 {
			if _, ok := seen[w.word]; !ok {
				seen[w.word] = struct{}{}
				results = append(results, w.word)
			}
		}
	}
	return results
}

func (m *MultitokenSpeller) stopSearching(candidates []string, originalWord string) bool {
	for _, candidate := range candidates {
		if m != nil && m.IsException != nil && m.IsException(originalWord, candidate) {
			return true
		}
		if candidate == originalWord {
			return true
		}
	}
	for _, candidate := range candidates {
		// Java: convertToTitleCaseIteratingChars
		if candidate == strings.ToLower(candidate) &&
			tools.ConvertToTitleCaseIteratingChars(candidate) == originalWord {
			return true
		}
	}
	return false
}

// discardRunOnWords ports MultitokenSpeller.discardRunOnWords.
// Requires IsMisspelledToken (Java SpellingCheckRule); nil → false.
func (m *MultitokenSpeller) discardRunOnWords(underlinedError string) bool {
	if m == nil || m.IsMisspelledToken == nil {
		return false
	}
	parts := splitBySpace(underlinedError)
	if len(parts) != 2 {
		return false
	}
	if tools.IsCapitalizedWord(parts[1]) {
		return false
	}
	if parts[0] == "" || parts[1] == "" {
		return true
	}
	// sugg1a + sugg1b: last UTF-16 unit of first token moved to second (Java substring)
	u0 := utf16.Encode([]rune(parts[0]))
	if len(u0) == 0 {
		return true
	}
	sugg1a := string(utf16.Decode(u0[:len(u0)-1]))
	sugg1b := string(utf16.Decode(u0[len(u0)-1:])) + parts[1]
	if !m.IsMisspelledToken(sugg1a) && !m.IsMisspelledToken(sugg1b) {
		return true
	}
	// sugg2a + sugg2b: first UTF-16 unit of second moved to first (Java charAt(0)/substring(1))
	u1 := utf16.Encode([]rune(parts[1]))
	if len(u1) == 0 {
		return true
	}
	sugg2a := parts[0] + string(utf16.Decode(u1[:1]))
	sugg2b := string(utf16.Decode(u1[1:]))
	return !m.IsMisspelledToken(sugg2a) && !m.IsMisspelledToken(sugg2b)
}

// distancesPerWord ports MultitokenSpeller.distancesPerWord.
func distancesPerWord(parts1, parts2 []string, s1, s2 string) []int {
	if len(parts1) == len(parts2) && len(parts1) > 1 {
		out := make([]int, len(parts1))
		for i := range parts1 {
			out[i] = levenshteinDistance(parts1[i], parts2[i])
		}
		return out
	}
	return []int{levenshteinDistance(s1, s2)}
}

// maxEditDistance ports MultitokenSpeller.maxEditDistance.
func maxEditDistance(normalizedCandidate, normalizedWord string) int {
	totalLength := utf16Len(normalizedWord)
	correctLength := totalLength - numberOfCorrectChars(normalizedCandidate, normalizedWord)
	firstCharWrong := float64(0)
	for _, d := range firstCharacterDistances(normalizedCandidate, normalizedWord) {
		firstCharWrong += d
	}
	if correctLength <= 7 {
		return int(2 - firstCharWrong)
	}
	return int(2 + 0.25*float64(correctLength-7) - 0.6*firstCharWrong)
}

func firstCharacterDistances(s1, s2 string) []float64 {
	parts1 := splitBySpace(s1)
	parts2 := splitBySpace(s2)
	// for now, only phrase with two tokens
	if len(parts1) == len(parts2) && len(parts1) == 2 {
		out := make([]float64, 2)
		for i := 0; i < 2; i++ {
			if parts1[i] == "" || parts2[i] == "" {
				out[i] = 1
				continue
			}
			// Java charAt(0) — first UTF-16 code unit
			r1 := rune(utf16.Encode([]rune(parts1[i]))[0])
			r2 := rune(utf16.Encode([]rune(parts2[i]))[0])
			out[i] = charDistance(r1, r2)
		}
		return out
	}
	return []float64{0}
}

func charDistance(a, b rune) float64 {
	if a == b {
		return 0
	}
	if (a == 's' && b == 'z') || (a == 'z' && b == 's') {
		return 0.2
	}
	if (a == 'b' && b == 'v') || (a == 'v' && b == 'b') {
		return 0.2
	}
	if (a == 'i' && b == 'y') || (a == 'y' && b == 'i') {
		return 0
	}
	return 1
}

func numberOfCorrectChars(s1, s2 string) int {
	parts1 := strings.Split(s1, " ")
	parts2 := strings.Split(s2, " ")
	correct := 0
	if len(parts1) == len(parts2) && len(parts1) > 1 {
		for i := range parts1 {
			if parts1[i] == parts2[i] {
				correct += utf16Len(parts1[i])
			}
		}
	}
	return correct
}

func splitBySpace(s string) []string {
	// Java StringUtils.split(s, ' ') — only ASCII space, omit empty segments.
	// (Not strings.Fields — that also splits on tabs/newlines/other unicode spaces.)
	if s == "" {
		return nil
	}
	raw := strings.Split(s, " ")
	out := make([]string, 0, len(raw))
	for _, p := range raw {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getNormalizeKey(word string) string {
	// Java: removeDiacritics(word.toLowerCase()).replace("-", " ") — no collapse
	s := tools.RemoveDiacritics(strings.ToLower(word))
	return strings.ReplaceAll(s, "-", " ")
}

// utf16Len ports Java String.length() for length comparisons in MultitokenSpeller.
func utf16Len(s string) int {
	return len([]uint16(utf16.Encode([]rune(s))))
}

func collapseWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func addToMap(m map[string][]string, key, value string) {
	for _, v := range m[key] {
		if v == value {
			return
		}
	}
	m[key] = append(m[key], value)
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// levenshteinDistance ports MultitokenSpeller.levenshteinDistance
// (Levenshtein + normalizeSimilarChars min + anagram tweaks).
func levenshteinDistance(s1, s2 string) int {
	if strings.ReplaceAll(s1, " ", "") == strings.ReplaceAll(s2, " ", "") {
		return 0
	}
	distance := rawLevenshtein(s1, s2)
	ns1 := normalizeSimilarChars(s1)
	ns2 := normalizeSimilarChars(s2)
	if s1 != ns1 || s2 != ns2 {
		if d2 := rawLevenshtein(ns1, ns2); d2 < distance {
			distance = d2
		}
	}
	anagram := tools.IsAnagram(s1, s2)
	// consider transpositions without having a Damerau-Levenshtein method
	if distance > 1 && anagram {
		distance--
	}
	if distance > 0 && utf16Len(s1) == utf16Len(s2) && anagram {
		distance = 1
	}
	return distance
}

func normalizeSimilarChars(s string) string {
	// Java: s.replace("y", "i").replace("ko", "co").replace("ka", "ca")
	s = strings.ReplaceAll(s, "y", "i")
	s = strings.ReplaceAll(s, "ko", "co")
	s = strings.ReplaceAll(s, "ka", "ca")
	return s
}

func rawLevenshtein(a, b string) int {
	// Apache commons-text LevenshteinDistance.apply(CharSequence) indexes with charAt
	// → UTF-16 code units (Java String), not Unicode code points.
	ua, ub := utf16.Encode([]rune(a)), utf16.Encode([]rune(b))
	la, lb := len(ua), len(ub)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	cur := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		cur[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if ua[i-1] == ub[j-1] {
				cost = 0
			}
			ins := cur[j-1] + 1
			del := prev[j] + 1
			sub := prev[j-1] + cost
			cur[j] = min3(ins, del, sub)
		}
		prev, cur = cur, prev
	}
	return prev[lb]
}

func min3(a, b, c int) int {
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
