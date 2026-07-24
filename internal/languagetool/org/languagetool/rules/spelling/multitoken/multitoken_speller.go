package multitoken

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/broker"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const maxLengthDiff = 3

// suggestionsCacheMaxSize ports Guava CacheBuilder.maximumSize(2000).
const suggestionsCacheMaxSize = 2000

// whitespaceAndSep ports MultitokenSpeller.WHITESPACE_AND_SEP = \p{Zs}+
// (Unicode space separators only — not tabs/newlines; Java Pattern.compile).
var whitespaceAndSep = regexp.MustCompile(`\p{Zs}+`)

// WeightedSuggestion maps org.languagetool.rules.spelling.morfologik.WeightedSuggestion
// (same fields/compare). Multitoken cannot import the morfologik package (import cycle:
// morfologik → spelling → patterns → multitoken). Field-identical pure mapping only.
type WeightedSuggestion struct {
	Word   string
	Weight int
}

// NewWeightedSuggestion ports morfologik.WeightedSuggestion constructor.
func NewWeightedSuggestion(word string, weight int) WeightedSuggestion {
	return WeightedSuggestion{Word: word, Weight: weight}
}

// GetWord ports getWord().
func (w WeightedSuggestion) GetWord() string { return w.Word }

// GetWeight ports getWeight().
func (w WeightedSuggestion) GetWeight() int { return w.Weight }

// sortWeightedByWeight ports Collections.sort(List<WeightedSuggestion>):
// stable ascending by weight only (Java Comparable).
func sortWeightedByWeight(s []WeightedSuggestion) {
	sort.SliceStable(s, func(i, j int) bool {
		return s[i].Weight < s[j].Weight
	})
}

// MultitokenSpeller ports org.languagetool.rules.spelling.multitoken.MultitokenSpeller.
//
// Java constructor takes Language + resource file paths, stores
// language.getDefaultSpellingRule() and calls initMultitokenSpeller.
// Go wires the same dependencies as hooks (no Language composite in this leaf):
//   - PrepareLine           → language.prepareLineForSpeller
//   - IsMisspelledToken     → SpellingCheckRule.isMisspelled (discardRunOnWords)
//   - IsException           → protected isException
//   - GetAdditionalSuggestions → protected getAdditionalSuggestions
//   - InitMultitokenSpeller → DataBroker resource streams + once-guard
type MultitokenSpeller struct {
	// suggestionsMap: first UTF-16 unit of normalized key → normalizedKey → original lines
	// (Java HashMap<Character, HashMap<String, List<String>>> suggestionsMap)
	suggestionsMap map[rune]map[string][]string
	// suggestionsMapNoSpacesKey: normalized key without spaces → original lines
	// (Java HashMap<String, List<String>> suggestionsMapNoSpacesKey)
	suggestionsMapNoSpacesKey map[string][]string

	// PrepareLine ports language.prepareLineForSpeller (default: identity single line).
	PrepareLine func(line string) []string
	// IsException ports protected MultitokenSpeller.isException (default false).
	IsException func(original, candidate string) bool
	// GetAdditionalSuggestions ports protected getAdditionalSuggestions (default empty).
	// If any additional word equals originalWord, computeSuggestions returns empty.
	GetAdditionalSuggestions func(originalWord string) []WeightedSuggestion
	// IsMisspelledToken ports SpellingCheckRule.isMisspelled for discardRunOnWords.
	// Nil: discardRunOnWords returns false (Java spellingRule null → NPE; leaf-safe hook
	// when base SpellingCheckRule is not wired yet — callers must set the real rule).
	IsMisspelledToken func(token string) bool

	mu    sync.RWMutex
	// suggestionsCache ports Guava Cache maximumSize(2000); key is whitespace-normalized word
	// only (Java does not include areTokensAcceptedBySpeller in the key).
	suggestionsCache map[string][]string
	// inited is the Java suggestionsMap != null once-guard for initMultitokenSpeller.
	inited bool
}

// NewMultitokenSpeller builds an empty MultitokenSpeller (maps ready; not yet loaded).
// Prefer InitMultitokenSpeller / LoadWords to fill the dictionary like Java.
func NewMultitokenSpeller() *MultitokenSpeller {
	return &MultitokenSpeller{
		suggestionsMap:            map[rune]map[string][]string{},
		suggestionsMapNoSpacesKey: map[string][]string{},
		suggestionsCache:          map[string][]string{},
		PrepareLine:               func(line string) []string { return []string{line} },
		// inited false until InitMultitokenSpeller or first LoadWords that marks maps live.
		// Java constructs empty maps inside init; LoadWords path fills without once-guard
		// per-stream (Java loop body). InitMultitokenSpeller applies the once-guard.
	}
}

// NewMultitokenSpellerFromResources ports MultitokenSpeller(Language, List<String>):
// prepareLine + isMisspelled hooks + DataBroker resource paths loaded once.
func NewMultitokenSpellerFromResources(
	prepareLine func(line string) []string,
	isMisspelled func(token string) bool,
	b broker.ResourceDataBroker,
	filePaths []string,
) (*MultitokenSpeller, error) {
	m := NewMultitokenSpeller()
	if prepareLine != nil {
		m.PrepareLine = prepareLine
	}
	m.IsMisspelledToken = isMisspelled
	if err := m.InitMultitokenSpeller(b, filePaths); err != nil {
		return nil, err
	}
	return m, nil
}

// InitMultitokenSpeller ports initMultitokenSpeller(List<String> filePaths).
// Once-guard: if already initialized, returns nil without reloading (Java early return).
// Loads each path via broker.GetFromResourceDirAsStream (Java JLanguageTool.getDataBroker()).
func (m *MultitokenSpeller) InitMultitokenSpeller(b broker.ResourceDataBroker, filePaths []string) error {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	if m.inited {
		m.mu.Unlock()
		return nil
	}
	// Java: suggestionsMap = new HashMap<>(); suggestionsMapNoSpacesKey = new HashMap<>();
	m.suggestionsMap = map[rune]map[string][]string{}
	m.suggestionsMapNoSpacesKey = map[string][]string{}
	m.mu.Unlock()

	if len(filePaths) > 0 && b == nil {
		return fmt.Errorf("MultitokenSpeller.initMultitokenSpeller: DataBroker is nil (Java JLanguageTool.getDataBroker())")
	}
	for _, filePath := range filePaths {
		stream, err := b.GetFromResourceDirAsStream(filePath)
		if err != nil {
			// Java wraps IOException in RuntimeException
			return err
		}
		err = m.loadWordsFromReader(stream)
		_ = stream.Close()
		if err != nil {
			return err
		}
	}
	m.mu.Lock()
	m.inited = true
	m.mu.Unlock()
	return nil
}

// LoadWords loads multiword dictionary lines from r (skip #/empty, strip trailing comments).
// Single-token lines are ignored (same as Java). Used by language loaders for each file
// stream (Java for-loop body over filePaths). Does not apply the once-guard so multiple
// files can be appended like the Java for-loop inside a single init.
func (m *MultitokenSpeller) LoadWords(r io.Reader) error {
	return m.loadWordsFromReader(r)
}

func (m *MultitokenSpeller) loadWordsFromReader(r io.Reader) error {
	sc := bufio.NewScanner(r)
	// Java BufferedReader can handle long lines; raise Go default 64K if needed.
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		lineOriginal := sc.Text()
		if lineOriginal == "" || lineOriginal[0] == '#' { // ignore comments
			continue
		}
		// Java: StringUtils.substringBefore(lineOriginal, "#").trim()
		if i := strings.Index(lineOriginal, "#"); i >= 0 {
			lineOriginal = lineOriginal[:i]
		}
		lineOriginal = tools.JavaStringTrim(lineOriginal)
		prep := m.PrepareLine
		if prep == nil {
			prep = func(line string) []string { return []string{line} }
		}
		// Java: for (String line : language.prepareLineForSpeller(...))
		for _, line := range prep(lineOriginal) {
			if line == "" {
				continue
			}
			normalizedKey := getNormalizeKey(line)
			if !strings.Contains(normalizedKey, " ") {
				// Ignore one-token suggestions
				continue
			}
			// Java: Character firstChar = normalizedKey.charAt(0) — UTF-16 unit
			firstChar := firstUTF16UnitAsRune(normalizedKey)
			m.mu.Lock()
			if m.suggestionsMap[firstChar] == nil {
				m.suggestionsMap[firstChar] = map[string][]string{}
			}
			addToMap(m.suggestionsMap[firstChar], normalizedKey, line)
			addToMap(m.suggestionsMapNoSpacesKey, strings.ReplaceAll(normalizedKey, " ", ""), line)
			m.mu.Unlock()
		}
	}
	return sc.Err()
}

// GetSuggestions ports getSuggestions(String) → getSuggestions(originalWord, false).
func (m *MultitokenSpeller) GetSuggestions(originalWord string) []string {
	return m.GetSuggestionsOpts(originalWord, false)
}

// GetSuggestionsOpts ports getSuggestions(String originalWord, boolean areTokensAcceptedBySpeller).
// Collapses \p{Zs}+ then cache-gets computeSuggestions (key is only the normalized word).
func (m *MultitokenSpeller) GetSuggestionsOpts(originalWord string, areTokensAcceptedBySpeller bool) []string {
	originalWord = collapseWhitespace(originalWord)
	normalizedWord := originalWord
	accepted := areTokensAcceptedBySpeller

	m.mu.RLock()
	if cached, ok := m.suggestionsCache[normalizedWord]; ok {
		m.mu.RUnlock()
		return append([]string(nil), cached...)
	}
	m.mu.RUnlock()

	results := m.computeSuggestions(normalizedWord, accepted)

	m.mu.Lock()
	// approximate maximumSize(2000): drop one arbitrary entry when full
	if len(m.suggestionsCache) >= suggestionsCacheMaxSize {
		for k := range m.suggestionsCache {
			delete(m.suggestionsCache, k)
			break
		}
	}
	// cache a copy so later mutation of returned slice cannot poison cache
	m.suggestionsCache[normalizedWord] = append([]string(nil), results...)
	m.mu.Unlock()
	return append([]string(nil), results...)
}

// computeSuggestions ports private computeSuggestions.
func (m *MultitokenSpeller) computeSuggestions(originalWord string, areTokensAcceptedBySpeller bool) []string {
	// Java: word = originalWord.replace("- ", "-").replace(" -", "-");
	word := strings.ReplaceAll(strings.ReplaceAll(originalWord, "- ", "-"), " -", "-")
	// Java discardRunOnWords(word) — may throw IOException; Go returns empty on true
	if m.discardRunOnWords(word) {
		return []string{}
	}
	normalizedWord := getNormalizeKey(word)
	var weightedCandidates []WeightedSuggestion

	normalizedWordNoSpaces := strings.ReplaceAll(normalizedWord, " ", "")
	m.mu.RLock()
	if candidates, ok := m.suggestionsMapNoSpacesKey[normalizedWordNoSpaces]; ok {
		if m.stopSearching(candidates, originalWord) {
			m.mu.RUnlock()
			return []string{}
		}
		for _, candidate := range candidates {
			weightedCandidates = append(weightedCandidates, NewWeightedSuggestion(candidate, 0))
		}
	}
	// Java: Character firstChar = normalizedWord.charAt(0)
	// Empty normalizedWord: Java throws StringIndexOutOfBoundsException; Go returns empty
	// (no invent of suggestions — only avoids panic on empty input).
	if normalizedWord != "" {
		firstChar := firstUTF16UnitAsRune(normalizedWord)
		byChar := m.suggestionsMap[firstChar]
		if len(weightedCandidates) == 0 && byChar != nil {
			for normalizedCandidate, candidates := range byChar {
				if m.stopSearching(candidates, originalWord) {
					m.mu.RUnlock()
					return []string{}
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
					for _, candidate := range candidates {
						weightedCandidates = append(weightedCandidates, NewWeightedSuggestion(candidate, 0))
					}
					// "continue" allows several candidates with different casing
					if len(weightedCandidates) == 2 {
						break
					}
					continue
				}
				// for very short candidates, allow only distance=0 (casing and diacritics differences)
				if utf16Len(normalizedCandidate) < 7 {
					continue
				}
				exceedsMaxDistancePerToken := false
				for i := 0; i < len(distances); i++ {
					// usually 2, but 1 for short words
					maxDistance := 1
					if i < len(wordParts) && i < len(candidateParts) &&
						utf16Len(wordParts[i]) > 5 && utf16Len(candidateParts[i]) > 4 {
						maxDistance = 2
					}
					if distances[i] > maxDistance {
						exceedsMaxDistancePerToken = true
						break
					}
				}
				if exceedsMaxDistancePerToken {
					continue
				}
				if totalDistance <= maxEditDistance(normalizedCandidate, normalizedWord) {
					for _, candidate := range candidates {
						weightedCandidates = append(weightedCandidates, NewWeightedSuggestion(candidate, totalDistance))
					}
				}
			}
		}
	}
	m.mu.RUnlock()

	// Java: for (WeightedSuggestion additionalSuggestion : getAdditionalSuggestions(word))
	if m.GetAdditionalSuggestions != nil {
		for _, additionalSuggestion := range m.GetAdditionalSuggestions(word) {
			if additionalSuggestion.GetWord() != originalWord {
				weightedCandidates = append(weightedCandidates, additionalSuggestion)
			} else {
				return []string{}
			}
		}
	}

	if len(weightedCandidates) == 0 {
		return []string{}
	}
	// Java: Collections.sort(weightedCandidates) — stable ascending by weight
	sortWeightedByWeight(weightedCandidates)

	weightFirstCandidate := weightedCandidates[0].GetWeight()
	// Java: weightedCandidates.get(0).getWord().toUpperCase().equals(originalWord)
	if areTokensAcceptedBySpeller && strings.ToUpper(weightedCandidates[0].GetWord()) == originalWord {
		// don't correct all-upper case words accepted by the speller
		return []string{}
	}
	if areTokensAcceptedBySpeller && weightFirstCandidate > 1 {
		return []string{}
	}
	var results []string
	for _, weightedCandidate := range weightedCandidates {
		// keep only candidates with the distance of the first candidate
		if weightedCandidate.GetWeight()-weightFirstCandidate < 1 && !containsString(results, weightedCandidate.GetWord()) {
			results = append(results, weightedCandidate.GetWord())
		}
	}
	return results
}

func containsString(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

// stopSearching ports MultitokenSpeller.stopSearching.
func (m *MultitokenSpeller) stopSearching(candidates []string, originalWord string) bool {
	for _, candidate := range candidates {
		if m.isException(originalWord, candidate) {
			return true
		}
		if candidate == originalWord {
			return true
		}
	}
	for _, candidate := range candidates {
		// Java: candidate.equals(candidate.toLowerCase()) && titleCase(candidate).equals(originalWord)
		if candidate == strings.ToLower(candidate) &&
			tools.ConvertToTitleCaseIteratingChars(candidate) == originalWord {
			return true
		}
	}
	return false
}

// isException ports protected isException (default false) via IsException hook.
func (m *MultitokenSpeller) isException(original, candidate string) bool {
	if m != nil && m.IsException != nil {
		return m.IsException(original, candidate)
	}
	return false
}

// discardRunOnWords ports MultitokenSpeller.discardRunOnWords.
// Requires IsMisspelledToken (Java SpellingCheckRule.isMisspelled). Nil hook → false.
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
		// probably emojis, or mal-formed chars
		return true
	}
	// Java: parts[0].substring(0, length-1) / substring(length-1) — UTF-16 units
	u0 := utf16.Encode([]rune(parts[0]))
	if len(u0) == 0 {
		return true
	}
	sugg1a := string(utf16.Decode(u0[:len(u0)-1]))
	sugg1b := string(utf16.Decode(u0[len(u0)-1:])) + parts[1]
	if !m.IsMisspelledToken(sugg1a) && !m.IsMisspelledToken(sugg1b) {
		return true
	}
	// Java: parts[0] + parts[1].charAt(0) / parts[1].substring(1) — UTF-16
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
// Java casts float results to int (truncate toward zero).
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

// firstCharacterDistances ports MultitokenSpeller.firstCharacterDistances.
func firstCharacterDistances(s1, s2 string) []float64 {
	parts1 := splitBySpace(s1)
	parts2 := splitBySpace(s2)
	// for now, only phrase with two tokens
	if len(parts1) == len(parts2) && len(parts1) == 2 {
		out := make([]float64, 2)
		for i := 0; i < 2; i++ {
			// Java charAt(0) — first UTF-16 code unit (empty part would NPE in Java)
			u1 := utf16.Encode([]rune(parts1[i]))
			u2 := utf16.Encode([]rune(parts2[i]))
			if len(u1) == 0 || len(u2) == 0 {
				// unreachable with StringUtils.split (omits empties); keep fail-closed
				out[i] = 1
				continue
			}
			out[i] = charDistance(rune(u1[0]), rune(u2[0]))
		}
		return out
	}
	return []float64{0}
}

// charDistance ports MultitokenSpeller.charDistance.
func charDistance(a, b rune) float64 {
	//TODO: full keyboard distances (same TODO as Java)
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

// numberOfCorrectChars ports MultitokenSpeller.numberOfCorrectChars.
// Java uses String.split(" ") (trailing empties discarded), not StringUtils.split.
func numberOfCorrectChars(s1, s2 string) int {
	parts1 := javaSplitOnSpace(s1)
	parts2 := javaSplitOnSpace(s2)
	correctTokens := 0
	if len(parts1) == len(parts2) && len(parts1) > 1 {
		for i := 0; i < len(parts1); i++ {
			if parts1[i] == parts2[i] {
				correctTokens += utf16Len(parts1[i])
			}
		}
	}
	return correctTokens
}

// splitBySpace ports MultitokenSpeller.splitBySpace → StringUtils.split(s, ' ').
// Only ASCII space; omits empty segments (unlike strings.Fields / String.split).
func splitBySpace(s string) []string {
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

// javaSplitOnSpace mirrors Java String.split(" ") (limit 0): trailing empties discarded,
// internal empties kept. Used by numberOfCorrectChars only.
func javaSplitOnSpace(s string) []string {
	parts := strings.Split(s, " ")
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

// getNormalizeKey ports MultitokenSpeller.getNormalizeKey:
// removeDiacritics(word.toLowerCase()).replace("-", " ")
func getNormalizeKey(word string) string {
	s := tools.RemoveDiacritics(strings.ToLower(word))
	return strings.ReplaceAll(s, "-", " ")
}

// firstUTF16UnitAsRune ports Java String.charAt(0) used as Character map key.
func firstUTF16UnitAsRune(s string) rune {
	if s == "" {
		return 0
	}
	u := utf16.Encode([]rune(s))
	if len(u) == 0 {
		return 0
	}
	return rune(u[0])
}

// utf16Len ports Java String.length() for MultitokenSpeller length comparisons.
func utf16Len(s string) int {
	return len(utf16.Encode([]rune(s)))
}

// collapseWhitespace ports getSuggestions:
// WHITESPACE_AND_SEP.matcher(originalWord).replaceAll(" ")
func collapseWhitespace(s string) string {
	return whitespaceAndSep.ReplaceAllString(s, " ")
}

// addToMap ports MultitokenSpeller.addToMap.
func addToMap(m map[string][]string, key, value string) {
	list := m[key]
	for _, v := range list {
		if v == value {
			return
		}
	}
	m[key] = append(list, value)
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// levenshteinDistance ports MultitokenSpeller.levenshteinDistance
// (Apache Levenshtein + normalizeSimilarChars min + anagram tweaks).
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
	// Java: s1.length()==s2.length() — UTF-16
	if distance > 0 && utf16Len(s1) == utf16Len(s2) && anagram {
		distance = 1
	}
	return distance
}

// normalizeSimilarChars ports MultitokenSpeller.normalizeSimilarChars.
func normalizeSimilarChars(s string) string {
	// Java: s.replace("y", "i").replace("ko", "co").replace("ka", "ca")
	s = strings.ReplaceAll(s, "y", "i")
	s = strings.ReplaceAll(s, "ko", "co")
	s = strings.ReplaceAll(s, "ka", "ca")
	return s
}

// rawLevenshtein ports Apache commons-text LevenshteinDistance.apply(CharSequence)
// which indexes with charAt → UTF-16 code units.
func rawLevenshtein(a, b string) int {
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
