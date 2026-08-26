package morfologik

import (
	"strings"
	"unicode/utf16"
)

// OrderedStringListMap ports Java LinkedHashMap<String, List<String>>.
type OrderedStringListMap struct {
	Keys []string
	M    map[string][]string
}

// Len is nil-safe key count.
func (m *OrderedStringListMap) Len() int {
	if m == nil || m.M == nil {
		return 0
	}
	return len(m.M)
}

// Add appends val; first insert records key order.
func (m *OrderedStringListMap) Add(key, val string) {
	if m.M == nil {
		m.M = map[string][]string{}
	}
	if _, ok := m.M[key]; !ok {
		m.Keys = append(m.Keys, key)
	}
	m.M[key] = append(m.M[key], val)
}

// Get returns values for key.
func (m *OrderedStringListMap) Get(key string) []string {
	if m == nil || m.M == nil {
		return nil
	}
	return m.M[key]
}

func parseConversionPairsInfo(value string) [][2]string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	seen := map[string]struct{}{}
	var out [][2]string
	for _, part := range splitCommaParts(value) {
		k, v, ok := splitTwoFields(part)
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

func parseReplacementPairsInfo(value string) []ReplPair {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	var out []ReplPair
	for _, part := range splitCommaParts(value) {
		k, v, ok := splitTwoFields(part)
		if !ok {
			continue
		}
		// Java: '_' → space (hunspell REP)
		k = strings.ReplaceAll(k, "_", " ")
		v = strings.ReplaceAll(v, "_", " ")
		out = append(out, ReplPair{From: k, To: v})
	}
	return out
}

func parseEquivalentCharsInfo(value string) map[rune][]rune {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	out := map[rune][]rune{}
	for _, part := range splitCommaParts(value) {
		fields := strings.Fields(part)
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

func partitionReplPairsInfo(pairs []ReplPair) (short []ReplPair, theRest *OrderedStringListMap) {
	theRest = &OrderedStringListMap{}
	for _, p := range pairs {
		toLen := len(utf16.Encode([]rune(p.To)))
		if toLen == 1 || toLen == 2 {
			short = append(short, p)
			continue
		}
		theRest.Add(p.From, p.To)
	}
	return short, theRest
}

func splitCommaParts(value string) []string {
	var out []string
	start := 0
	for i := 0; i < len(value); i++ {
		if value[i] == ',' {
			seg := strings.TrimSpace(value[start:i])
			if seg != "" {
				out = append(out, seg)
			}
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

func splitTwoFields(p string) (string, string, bool) {
	fields := strings.Fields(strings.TrimSpace(p))
	if len(fields) < 2 {
		return "", "", false
	}
	return fields[0], fields[1], true
}

// applyConversionPairs ports DictionaryLookup.applyReplacements (ordered pairs).
func applyConversionPairs(word string, pairs [][2]string) string {
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
			from = idx + len(key)
		}
	}
	return s
}
