package morfologik

import (
	"regexp"
	"strings"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// spacePlusRE ports Java stringPair.split(" +") — one or more ASCII spaces only.
var spacePlusRE = regexp.MustCompile(` +`)

// MAX_RECURSION_LEVEL ports morfologik Speller.MAX_RECURSION_LEVEL for getAllReplacements.
const maxReplacementRecursion = 6

// UPPER_SEARCH_LIMIT ports Speller.UPPER_SEARCH_LIMIT for replacement variant edit searches.
const upperSearchLimit = 15

// ApplyConversionPairs ports DictionaryLookup.applyReplacements (ordered LinkedHashMap).
// Empty → word unchanged. Bug-for-bug: after each replace, next search starts at index+len(key).
func ApplyConversionPairs(word string, pairs [][2]string) string {
	if word == "" || len(pairs) == 0 {
		return word
	}
	s := word
	for _, p := range pairs {
		key, val := p[0], p[1]
		if key == "" {
			continue
		}
		from := 0
		for from <= len(s) {
			rel := strings.Index(s[from:], key)
			if rel < 0 {
				break
			}
			idx := from + rel
			s = s[:idx] + val + s[idx+len(key):]
			// Java: index = sb.indexOf(key, index + key.length()) after replace
			from = idx + len(key)
		}
	}
	return s
}

// ParseConversionPairs ports DictionaryAttribute.INPUT_CONVERSION / OUTPUT_CONVERSION fromString.
// Format: "a b, c d" → ordered pairs; first value wins per key.
func ParseConversionPairs(value string) [][2]string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parts := splitCommaPairs(value)
	seen := map[string]struct{}{}
	var out [][2]string
	for _, p := range parts {
		k, v, ok := splitSpacePair(p)
		if !ok {
			continue
		}
		if _, dup := seen[k]; dup {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, [2]string{k, v})
	}
	return out
}

// ReplacementPair is one from→to entry; multiple targets per source allowed (Java List).
type ReplacementPair struct {
	From string // may include ^ / $ anchors
	To   string
}

// ParseReplacementPairs ports DictionaryAttribute.REPLACEMENT_PAIRS fromString.
// Format: "a b, a c, x y" → multi-map (same key multiple values).
// Java: '_' represents a space (hunspell REP convention).
func ParseReplacementPairs(value string) []ReplacementPair {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parts := splitCommaPairs(value)
	var out []ReplacementPair
	for _, p := range parts {
		k, v, ok := splitSpacePair(p)
		if !ok {
			continue
		}
		// Java: twoStrings[0].replace('_', ' ') / twoStrings[1].replace('_', ' ')
		k = strings.ReplaceAll(k, "_", " ")
		v = strings.ReplaceAll(v, "_", " ")
		out = append(out, ReplacementPair{From: k, To: v})
	}
	return out
}

func splitCommaPairs(value string) []string {
	// Java: value.split(",\\s*")
	var out []string
	start := 0
	for i := 0; i < len(value); i++ {
		if value[i] == ',' {
			seg := strings.TrimSpace(value[start:i])
			if seg != "" {
				out = append(out, seg)
			}
			// skip following spaces
			j := i + 1
			for j < len(value) && (value[j] == ' ' || value[j] == '\t') {
				j++
			}
			start = j
			i = j - 1
		}
	}
	if start < len(value) {
		seg := strings.TrimSpace(value[start:])
		if seg != "" {
			out = append(out, seg)
		}
	}
	return out
}

func splitSpacePair(p string) (string, string, bool) {
	// Java: stringPair.split(" +") — one or more ASCII spaces (not Unicode Fields).
	p = tools.JavaStringTrim(p)
	if p == "" {
		return "", "", false
	}
	fields := spacePlusRE.Split(p, -1)
	// drop empties from leading/trailing
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if f != "" {
			out = append(out, f)
		}
	}
	if len(out) < 2 {
		return "", "", false
	}
	// Java uses twoStrings[0] and twoStrings[1] only
	return out[0], out[1], true
}

// ParseEquivalentChars ports DictionaryAttribute.EQUIVALENT_CHARS fromString.
// Format: "x ź, l ł, u ó" → map[x]=[ź], map[l]=[ł], ...
func ParseEquivalentChars(value string) map[rune][]rune {
	value = tools.JavaStringTrim(value)
	if value == "" {
		return nil
	}
	out := map[rune][]rune{}
	for _, part := range splitCommaPairs(value) {
		// Java: part split on spaces for "x ź" pairs — ASCII space+ only.
		fields := spacePlusRE.Split(tools.JavaStringTrim(part), -1)
		// drop empties
		clean := fields[:0]
		for _, f := range fields {
			if f != "" {
				clean = append(clean, f)
			}
		}
		fields = clean
		if len(fields) != 2 {
			continue
		}
		fr := []rune(fields[0])
		tr := []rune(fields[1])
		if len(fr) != 1 || len(tr) != 1 {
			continue
		}
		out[fr[0]] = append(out[fr[0]], tr[0])
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func isStartAnchored(key string) bool { return strings.HasPrefix(key, "^") }
func isEndAnchored(key string) bool   { return strings.HasSuffix(key, "$") }

func stripAnchors(key string) string {
	start := 0
	end := len(key)
	if strings.HasPrefix(key, "^") {
		start = 1
	}
	if strings.HasSuffix(key, "$") {
		end--
	}
	if start >= end {
		return ""
	}
	return key[start:end]
}

// LinkedHashStringListMap ports Java LinkedHashMap<String, List<String>> (insertion-ordered keys).
type LinkedHashStringListMap struct {
	Keys []string
	M    map[string][]string
}

// Len returns the number of keys (nil-safe).
func (m *LinkedHashStringListMap) Len() int {
	if m == nil || m.M == nil {
		return 0
	}
	return len(m.M)
}

// Add appends val to key's list; first insert records key order.
func (m *LinkedHashStringListMap) Add(key, val string) {
	if m.M == nil {
		m.M = map[string][]string{}
	}
	if _, ok := m.M[key]; !ok {
		m.Keys = append(m.Keys, key)
	}
	m.M[key] = append(m.M[key], val)
}

// Get returns the list for key (nil if absent).
func (m *LinkedHashStringListMap) Get(key string) []string {
	if m == nil || m.M == nil {
		return nil
	}
	return m.M[key]
}

// partitionReplacementPairs splits pairs like Java Speller.createReplacementsMaps:
// target len 1 / 2 → shortPairs (loaded into SpellerFSA anyToOne/anyToTwo for findRepl);
// longer → theRest for getAllReplacements (LinkedHashMap order).
func partitionReplacementPairs(pairs []ReplacementPair) (theRest *LinkedHashStringListMap, short []ReplacementPair) {
	theRest = &LinkedHashStringListMap{}
	for _, p := range pairs {
		toRunes := []rune(p.To)
		// Java: s.length() is UTF-16
		toLen := len(utf16.Encode(toRunes))
		if toLen == 1 || toLen == 2 {
			short = append(short, p)
			continue
		}
		theRest.Add(p.From, p.To)
	}
	return theRest, short
}

// GetAllReplacements ports Speller.getAllReplacements (theRest multi-char targets only).
// Iterates keys in LinkedHashMap insertion order (Java DictionaryMetadata.replacementPairs).
func GetAllReplacements(str string, theRest *LinkedHashStringListMap, fromIndex, level int) []string {
	if theRest == nil || theRest.Len() == 0 {
		return []string{str}
	}
	if level > maxReplacementRecursion {
		return []string{str}
	}
	sb := str
	index := 120 // MAX_WORD_LENGTH stand-in (sentinel)
	key := ""
	keyLength := 0
	found := false
	strippedKeyForSelected := ""
	for _, auxKey := range theRest.Keys {
		startAnchor := isStartAnchored(auxKey)
		endAnchor := isEndAnchored(auxKey)
		stripped := auxKey
		if startAnchor || endAnchor {
			stripped = stripAnchors(auxKey)
		}
		auxIndex := -1
		if startAnchor && fromIndex > 0 {
			continue
		} else if startAnchor {
			if strings.HasPrefix(sb, stripped) {
				auxIndex = 0
			}
		} else if endAnchor {
			expectedIndex := len(sb) - len(stripped)
			if expectedIndex >= fromIndex && expectedIndex >= 0 && strings.HasSuffix(sb, stripped) {
				auxIndex = expectedIndex
			}
		} else {
			if i := strings.Index(sb[fromIndex:], auxKey); i >= 0 {
				auxIndex = fromIndex + i
			}
			// Java uses auxKey not stripped for non-anchor indexOf
		}
		if auxIndex > -1 && (auxIndex < index || (auxIndex == index && !(len(stripped) < keyLength))) {
			index = auxIndex
			key = auxKey
			keyLength = len(stripped)
			strippedKeyForSelected = stripped
		}
	}
	var replaced []string
	if index < 120 {
		for _, rep := range theRest.Get(key) {
			if !found {
				replaced = append(replaced, GetAllReplacements(str, theRest, index+len(strippedKeyForSelected), level+1)...)
				found = true
			}
			// avoid unnecessary replacements
			ind := -1
			searchFrom := fromIndex - len(rep) + 1
			if searchFrom < 0 {
				searchFrom = 0
			}
			if i := strings.Index(sb[searchFrom:], rep); i >= 0 {
				ind = searchFrom + i
			}
			if len(rep) > len(strippedKeyForSelected) && ind > -1 &&
				(ind == index || ind == index-len(rep)+1) {
				continue
			}
			// branch with replacement
			newStr := sb[:index] + rep + sb[index+len(strippedKeyForSelected):]
			replaced = append(replaced, GetAllReplacements(newStr, theRest, index+len(rep), level+1)...)
		}
	}
	if !found {
		replaced = append(replaced, sb)
	}
	return replaced
}
