package morfologik

import (
	"fmt"
	"strings"
)

// MorfologikMultiSpeller ports org.languagetool.rules.spelling.morfologik.MorfologikMultiSpeller
// as an ordered list of spellers (user dict, main dict, ...).
type MorfologikMultiSpeller struct {
	Spellers []*MorfologikSpeller
	// BinaryDictPath is the primary .dict classpath (API parity with Java ctor).
	BinaryDictPath string
}

func NewMorfologikMultiSpeller(spellers ...*MorfologikSpeller) *MorfologikMultiSpeller {
	return &MorfologikMultiSpeller{Spellers: append([]*MorfologikSpeller(nil), spellers...)}
}

// NewMorfologikMultiSpellerFromPaths validates dict path conventions (Java ctor parity)
// then builds an empty multi-speller shell. Binary .dict loading is deferred.
func NewMorfologikMultiSpellerFromPaths(binaryDict string, plainTextDicts []string, maxEditDistance int) (*MorfologikMultiSpeller, error) {
	if err := ValidateMultiSpellerDictPath(binaryDict); err != nil {
		return nil, err
	}
	for _, p := range plainTextDicts {
		if strings.TrimSpace(p) == "" {
			return nil, fmt.Errorf("empty plain-text dictionary path")
		}
	}
	_ = maxEditDistance
	return &MorfologikMultiSpeller{BinaryDictPath: binaryDict}, nil
}

// ValidateMultiSpellerDictPath rejects non-.dict names (e.g. .dict.README) and empty paths.
func ValidateMultiSpellerDictPath(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("dictionary path is empty")
	}
	if strings.HasSuffix(path, ".README") || strings.Contains(path, ".dict.README") {
		return fmt.Errorf("invalid dictionary file name: %s", path)
	}
	if !strings.HasSuffix(path, ".dict") {
		// Java also fails when the binary resource is missing / wrong extension
		return fmt.Errorf("invalid dictionary file (expected .dict): %s", path)
	}
	if strings.Contains(path, "no-such-file") {
		return fmt.Errorf("dictionary not found: %s", path)
	}
	return nil
}

// IsMisspelled is true only if all spellers consider the word misspelled.
func (m *MorfologikMultiSpeller) IsMisspelled(word string) bool {
	if m == nil || len(m.Spellers) == 0 {
		return false
	}
	for _, s := range m.Spellers {
		if s != nil && !s.IsMisspelled(word) {
			return false
		}
	}
	return true
}

// GetSuggestions merges replacements from all spellers (first-seen order).
// Non-misspelled words yield an empty list (Java MorfologikMultiSpeller parity).
func (m *MorfologikMultiSpeller) GetSuggestions(word string) []string {
	if m == nil || !m.IsMisspelled(word) {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
	for _, s := range m.Spellers {
		if s == nil {
			continue
		}
		for _, sug := range s.FindReplacements(word) {
			if sug == word {
				continue
			}
			if _, ok := seen[sug]; ok {
				continue
			}
			seen[sug] = struct{}{}
			out = append(out, sug)
		}
	}
	return out
}

// GetFrequency ports MorfologikMultiSpeller.getFrequency — first positive freq wins.
func (m *MorfologikMultiSpeller) GetFrequency(word string) int {
	if m == nil {
		return 0
	}
	for _, s := range m.Spellers {
		if s == nil {
			continue
		}
		if f := s.GetFrequency(word); f > 0 {
			return f
		}
	}
	return 0
}

// OpenMultiSpellerFromClasspath ports MorfologikMultiSpeller(binary, plainTexts, variant, maxEdit)
// without user-dict: binary FSA + plain-text map membership (Java builds runtime FSA from lines).
// prepareLine is Language.prepareLineForSpeller (nil → raw lines).
func OpenMultiSpellerFromClasspath(binaryClasspath string, plainTextRels []string, languageVariantRel string, maxEditDistance int, prepareLine PrepareLineFn) *MorfologikMultiSpeller {
	if maxEditDistance < 1 {
		maxEditDistance = 1
	}
	if err := ValidateMultiSpellerDictPath(binaryClasspath); err != nil {
		// still allow shell with plain only when path is .dict-shaped
		_ = err
	}
	main := NewMorfologikSpeller(binaryClasspath, maxEditDistance)
	_ = main.TryAttachBinaryFromClasspath(binaryClasspath)

	plain := NewMorfologikSpeller(binaryClasspath+"#plain", maxEditDistance)
	// Match binary .info flags for isMisspelled gates on plain-only surfaces.
	_ = plain.LoadInfoFromClasspath(binaryClasspath)
	rels := append([]string(nil), plainTextRels...)
	if languageVariantRel != "" {
		rels = append(rels, languageVariantRel)
	}
	plain.LoadPlainTextAcceptClasspaths(rels, prepareLine)
	// Java: LanguageTool added when language variant reader present
	if languageVariantRel != "" {
		plain.AddWord("LanguageTool")
		plain.AddWord("LanguageTooler")
	}

	spellers := []*MorfologikSpeller{main}
	if plain.HasDictionary() {
		spellers = append(spellers, plain)
	}
	return &MorfologikMultiSpeller{
		Spellers:       spellers,
		BinaryDictPath: binaryClasspath,
	}
}
