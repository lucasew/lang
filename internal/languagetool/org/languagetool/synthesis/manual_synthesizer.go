package synthesis

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ManualSynthesizer ports org.languagetool.synthesis.ManualSynthesizer (map-based).
type ManualSynthesizer struct {
	// key: lemma + "\x00" + posTag → inflected forms
	byLemmaPOS   map[string][]string
	possibleTags map[string]struct{}
}

func NewManualSynthesizer(r io.Reader) (*ManualSynthesizer, error) {
	mapping, err := loadSynthMapping(r)
	if err != nil {
		return nil, err
	}
	s := &ManualSynthesizer{
		byLemmaPOS:   mapping,
		possibleTags: map[string]struct{}{},
	}
	for k := range mapping {
		// extract pos after null
		if i := strings.IndexByte(k, 0); i >= 0 {
			s.possibleTags[k[i+1:]] = struct{}{}
		}
	}
	return s, nil
}

func loadSynthMapping(r io.Reader) (map[string][]string, error) {
	m := map[string][]string{}
	sc := bufio.NewScanner(r)
	separator := "\t"
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "#separatorRegExp=") {
			separator = strings.TrimPrefix(line, "#separatorRegExp=")
			continue
		}
		if tools.IsEmptyStr(line) || line[0] == '#' {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		parts := strings.Split(line, separator)
		if len(parts) != 3 {
			return nil, fmt.Errorf("Unknown line format when loading manual synthesizer dictionary, expected 3 parts separated by '%s', found %d: '%s'", separator, len(parts), line)
		}
		form := parts[0]
		lemma := parts[1]
		if form == lemma {
			form = lemma
		}
		posTag := parts[2]
		key := lemma + "\x00" + posTag
		m[key] = append(m[key], form)
	}
	return m, sc.Err()
}

// Lookup returns inflected forms for lemma+POS, or nil if none (Java null).
func (s *ManualSynthesizer) Lookup(lemma, posTag string) []string {
	if lemma == "" || posTag == "" {
		// Java: null lemma or posTag → null; empty string is not null
		// but testLookupNonExisting uses "" and expects null
		// assertNull(lookup("", "")) - empty strings are non-null in Java
		// Wait: assertNull(synthesizer.lookup("", "")); - empty string is not null in Java
		// Looking at Java: if (lemma == null || posTag == null) return null;
		// So empty string should look up and return null if not found
	}
	// For Go we use empty string; nil not representable for string params.
	// Use special: we need pointer for nullability - change signature to match tests
	key := lemma + "\x00" + posTag
	forms, ok := s.byLemmaPOS[key]
	if !ok || len(forms) == 0 {
		return nil
	}
	return append([]string(nil), forms...)
}

// LookupPtr accepts optional nil for lemma/posTag (Java null).
func (s *ManualSynthesizer) LookupPtr(lemma, posTag *string) []string {
	if lemma == nil || posTag == nil {
		return nil
	}
	return s.Lookup(*lemma, *posTag)
}

func (s *ManualSynthesizer) GetPossibleTags() map[string]struct{} {
	return s.possibleTags
}
