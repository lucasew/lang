package multitoken

import (
	"bufio"
	"io"
	"strings"
	"sync"
	"unicode"
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
	mu                       sync.RWMutex
	cache                    map[string][]string
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
	word := strings.ReplaceAll(strings.ReplaceAll(originalWord, "- ", "-"), " -", "-")
	if word == "" {
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
	first := []rune(normalizedWord)[0]
	byChar := m.byFirstChar[first]
	if len(weightedCandidates) == 0 && byChar != nil {
		for normalizedCandidate, candidates := range byChar {
			if m.stopSearching(candidates, originalWord) {
				m.mu.RUnlock()
				return nil
			}
			if abs(len(normalizedCandidate)-len(word)) > maxLengthDiff {
				continue
			}
			dist := levenshtein(normalizedCandidate, normalizedWord)
			// short candidates: only exact (dist 0 after normalize)
			if len(normalizedCandidate) < 7 && dist > 0 {
				continue
			}
			if dist <= maxEditDistance(normalizedCandidate, normalizedWord) {
				for _, c := range candidates {
					weightedCandidates = append(weightedCandidates, weighted{c, dist})
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
	// sort by weight
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
		if candidate == strings.ToLower(candidate) && titleCase(candidate) == originalWord {
			return true
		}
	}
	return false
}

func maxEditDistance(normalizedCandidate, normalizedWord string) int {
	totalLength := len(normalizedWord)
	correctLength := totalLength // simplified vs Java numberOfCorrectChars
	_ = normalizedCandidate
	if correctLength <= 7 {
		return 2
	}
	return 2 + int(0.25*float64(correctLength-7))
}

func getNormalizeKey(word string) string {
	s := strings.ToLower(word)
	s = removeDiacritics(s)
	s = strings.ReplaceAll(s, "-", " ")
	return collapseWhitespace(s)
}

func collapseWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func removeDiacritics(s string) string {
	// Lightweight surface: drop common combining marks and map a few precomposed chars.
	// Full NFD needs golang.org/x/text; this covers Romance diacritics used in multiword lists.
	var b strings.Builder
	for _, r := range s {
		switch r {
		case 'á', 'à', 'â', 'ä', 'ã', 'å':
			b.WriteByte('a')
		case 'é', 'è', 'ê', 'ë':
			b.WriteByte('e')
		case 'í', 'ì', 'î', 'ï':
			b.WriteByte('i')
		case 'ó', 'ò', 'ô', 'ö', 'õ':
			b.WriteByte('o')
		case 'ú', 'ù', 'û', 'ü':
			b.WriteByte('u')
		case 'ý', 'ÿ':
			b.WriteByte('y')
		case 'ç':
			b.WriteByte('c')
		case 'ñ':
			b.WriteByte('n')
		case 'Á', 'À', 'Â', 'Ä', 'Ã', 'Å':
			b.WriteByte('A')
		case 'É', 'È', 'Ê', 'Ë':
			b.WriteByte('E')
		case 'Í', 'Ì', 'Î', 'Ï':
			b.WriteByte('I')
		case 'Ó', 'Ò', 'Ô', 'Ö', 'Õ':
			b.WriteByte('O')
		case 'Ú', 'Ù', 'Û', 'Ü':
			b.WriteByte('U')
		case 'Ý':
			b.WriteByte('Y')
		case 'Ç':
			b.WriteByte('C')
		case 'Ñ':
			b.WriteByte('N')
		default:
			if unicode.Is(unicode.Mn, r) {
				continue
			}
			b.WriteRune(r)
		}
	}
	return b.String()
}

func addToMap(m map[string][]string, key, value string) {
	for _, v := range m[key] {
		if v == value {
			return
		}
	}
	m[key] = append(m[key], value)
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	rs := []rune(s)
	rs[0] = unicode.ToUpper(rs[0])
	return string(rs)
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func levenshtein(a, b string) int {
	if strings.ReplaceAll(a, " ", "") == strings.ReplaceAll(b, " ", "") {
		return 0
	}
	ra, rb := []rune(a), []rune(b)
	// classic DP
	la, lb := len(ra), len(rb)
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
			if ra[i-1] == rb[j-1] {
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
