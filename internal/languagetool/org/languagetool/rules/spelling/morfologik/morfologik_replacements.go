package morfologik

import (
	"strings"
	"unicode/utf16"
)

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
	// Java: stringPair.split(" +") — one or more spaces
	p = strings.TrimSpace(p)
	fields := strings.Fields(p)
	if len(fields) < 2 {
		return "", "", false
	}
	// join rest as value? Java uses twoStrings[0] and twoStrings[1] only
	return fields[0], fields[1], true
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

// partitionReplacementPairs splits pairs like Java createReplacementsMaps:
// target len 1 / 2 go to HMatrix maps (returned separately); longer → theRest for getAllReplacements.
func partitionReplacementPairs(pairs []ReplacementPair) (theRest map[string][]string) {
	theRest = map[string][]string{}
	for _, p := range pairs {
		toRunes := []rune(p.To)
		// Java: s.length() is UTF-16
		toLen := len(utf16.Encode(toRunes))
		if toLen == 1 || toLen == 2 {
			// HMatrix path — not used in getAllReplacements; still store under theRest for
			// approximate edit path when HMatrix missing? Java does NOT put them in theRest.
			// Keep only length >= 3 for getAllReplacements.
			continue
		}
		theRest[p.From] = append(theRest[p.From], p.To)
	}
	return theRest
}

// GetAllReplacements ports Speller.getAllReplacements (theRest multi-char targets only).
func GetAllReplacements(str string, theRest map[string][]string, fromIndex, level int) []string {
	if theRest == nil || len(theRest) == 0 {
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
	for auxKey := range theRest {
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
		for _, rep := range theRest[key] {
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
