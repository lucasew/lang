package synthesis

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ManualSynthesizer ports org.languagetool.synthesis.ManualSynthesizer.
// Storage is map-based (lemma+"\x00"+posTag → encoded forms) with the same
// loadMapping / encodeForm / decodeForm / lookup semantics as Java. The Java
// Int2Int hash table is an implementation detail that does not change
// visible results for exact lemma+POS keys (hash collisions only affect
// internal filtering that exact keys make unnecessary).
type ManualSynthesizer struct {
	// key: lemma + "\x00" + posTag → forms encoded by encodeManualSynthForm
	byLemmaPOS   map[string][]string
	possibleTags map[string]struct{}
}

// manualSynthSuffixMarker ports ManualSynthesizer.SUFFIX_MARKER.
const manualSynthSuffixMarker = "+"

const manualSynthDefaultSeparator = "\t"

// NewManualSynthesizer ports ManualSynthesizer(InputStream).
func NewManualSynthesizer(r io.Reader) (*ManualSynthesizer, error) {
	mapping, err := loadSynthMapping(r)
	if err != nil {
		return nil, err
	}
	// groupByHash + encodeForm: reject forms starting with SUFFIX_MARKER;
	// store encodeForm(lemma, word) for each form.
	encoded := make(map[string][]string, len(mapping))
	possibleTags := make(map[string]struct{})
	for key, forms := range mapping {
		i := strings.IndexByte(key, 0)
		if i < 0 {
			continue
		}
		lemma := key[:i]
		posTag := key[i+1:]
		possibleTags[posTag] = struct{}{}
		out := make([]string, 0, len(forms))
		for _, word := range forms {
			if strings.HasPrefix(word, manualSynthSuffixMarker) {
				// Java: UnsupportedOperationException("Words can't start with +")
				return nil, fmt.Errorf("Words can't start with %s", manualSynthSuffixMarker)
			}
			out = append(out, encodeManualSynthForm(lemma, word))
		}
		encoded[key] = out
	}
	return &ManualSynthesizer{
		byLemmaPOS:   encoded,
		possibleTags: possibleTags,
	}, nil
}

// loadSynthMapping ports ManualSynthesizer.loadMapping.
func loadSynthMapping(r io.Reader) (map[string][]string, error) {
	m := map[string][]string{}
	sc := bufio.NewScanner(r)
	// Java Scanner is not bound to a small fixed buffer.
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 10*1024*1024)
	separator := manualSynthDefaultSeparator
	for sc.Scan() {
		// Java: line = line.trim() (String.trim).
		line := tools.JavaStringTrim(sc.Text())
		if strings.HasPrefix(line, "#separatorRegExp=") {
			// Java: separator = line.replace("#separatorRegExp=", "")
			separator = strings.ReplaceAll(line, "#separatorRegExp=", "")
			// fall through: still starts with '#' → continue
		}
		if tools.IsEmptyStr(line) || line[0] == '#' {
			continue
		}
		// Java: StringUtils.substringBefore(line, "#").trim()
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = tools.JavaStringTrim(line[:i])
		}
		// Java: line.split(separator) — regex when set via #separatorRegExp=
		parts, err := splitManualSynthParts(line, separator)
		if err != nil {
			return nil, err
		}
		if len(parts) != 3 {
			return nil, fmt.Errorf("Unknown line format when loading manual synthesizer dictionary, expected 3 parts separated by '%s', found %d: '%s'", separator, len(parts), line)
		}
		form := parts[0]
		lemma := parts[1]
		if form == lemma {
			// Java: form = lemma (share interned reference)
			form = lemma
		}
		posTag := parts[2]
		key := lemma + "\x00" + posTag
		m[key] = append(m[key], form)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return m, nil
}

// splitManualSynthParts ports Java String.split(separator) for ManualSynthesizer.
// Default separator is "\t". #separatorRegExp= sets a regex pattern.
func splitManualSynthParts(line, separator string) ([]string, error) {
	if separator == "\t" || separator == "" {
		return strings.Split(line, "\t"), nil
	}
	// Java String.split(regex) with limit 0 drops trailing empties; invalid pattern → PatternSyntaxException.
	re, err := regexp.Compile(separator)
	if err != nil {
		return nil, err
	}
	parts := re.Split(line, -1)
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts, nil
}

// encodeManualSynthForm ports ManualSynthesizer.encodeForm.
func encodeManualSynthForm(lemma, word string) string {
	if len(word) > len(lemma) && strings.HasPrefix(word, lemma) {
		return manualSynthSuffixMarker + word[len(lemma):]
	}
	// Java: word.startsWith(lemma.substring(0, lemma.length() - 1))
	// lemma.length()==0 → StringIndexOutOfBoundsException
	if len(lemma) == 0 {
		panic("String index out of range: -1")
	}
	if len(word) >= len(lemma) && strings.HasPrefix(word, lemma[:len(lemma)-1]) {
		return manualSynthSuffixMarker + manualSynthSuffixMarker + word[len(lemma)-1:]
	}
	return word
}

// decodeManualSynthForm ports ManualSynthesizer.decodeForm.
func decodeManualSynthForm(lemma, word string) string {
	if strings.HasPrefix(word, manualSynthSuffixMarker) {
		// word.startsWith(SUFFIX_MARKER, SUFFIX_MARKER.length())
		if strings.HasPrefix(word[len(manualSynthSuffixMarker):], manualSynthSuffixMarker) {
			// Java: lemma.substring(0, lemma.length() - 1) + word.substring(2)
			if len(lemma) == 0 {
				panic("String index out of range: -1")
			}
			return lemma[:len(lemma)-1] + word[len(manualSynthSuffixMarker)*2:]
		}
		return lemma + word[len(manualSynthSuffixMarker):]
	}
	return word
}

// Lookup ports ManualSynthesizer.lookup(String, String).
// Returns nil for Java null (missing entry). Empty lemma/posTag are non-null
// keys and return nil when absent (Java testLookupNonExisting).
func (s *ManualSynthesizer) Lookup(lemma, posTag string) []string {
	if s == nil {
		return nil
	}
	// Java: if (lemma == null || posTag == null) return null;
	// Go strings cannot be null; use LookupPtr for explicit nulls.
	key := lemma + "\x00" + posTag
	forms, ok := s.byLemmaPOS[key]
	if !ok {
		return nil
	}
	// Java filters by lemma.equals / posTag.equals inside a hash bucket; with exact
	// keys every stored form matches. Decode each (encodeForm/decodeForm pair).
	result := make([]string, 0, len(forms))
	for _, word := range forms {
		result = append(result, decodeManualSynthForm(lemma, word))
	}
	return result
}

// LookupPtr accepts optional nil for lemma/posTag (Java null → null list).
func (s *ManualSynthesizer) LookupPtr(lemma, posTag *string) []string {
	if lemma == nil || posTag == nil {
		return nil
	}
	return s.Lookup(*lemma, *posTag)
}

// GetPossibleTags ports ManualSynthesizer.getPossibleTags.
func (s *ManualSynthesizer) GetPossibleTags() map[string]struct{} {
	if s == nil {
		return nil
	}
	return s.possibleTags
}
