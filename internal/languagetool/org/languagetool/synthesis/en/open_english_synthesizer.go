package en

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// OpenEnglishSynthesizerFromDir loads Java English synthesizer resources from a
// resource/en directory (english_synth.dict, english_tags.txt, added.txt,
// removed.txt, do-not-synthesize.txt). Missing optional manual files are skipped.
// Returns nil if the binary synth dict cannot be opened (fail-closed, no invent).
func OpenEnglishSynthesizerFromDir(resourceDir string) *EnglishSynthesizer {
	if resourceDir == "" {
		return nil
	}
	dictPath := filepath.Join(resourceDir, "english_synth.dict")
	lookup, err := synthesis.OpenMorfologikSynthLookup(dictPath)
	if err != nil || lookup == nil {
		return nil
	}
	manual, _ := synthesis.LoadManualSynthesizerFile(filepath.Join(resourceDir, "added.txt"))
	removal := loadMergedManual(
		filepath.Join(resourceDir, "removed.txt"),
		filepath.Join(resourceDir, "do-not-synthesize.txt"),
	)
	s := NewEnglishSynthesizer(manual)
	s.Lookup = lookup
	s.Removal = removal
	if tags, err := synthesis.LoadPossibleTagsFile(filepath.Join(resourceDir, "english_tags.txt")); err == nil && len(tags) > 0 {
		s.PossibleTags = tags
	}
	// Java BaseSynthesizer createNumberSpeller / createRomanNumberer
	s.SorFileName = EnglishSorFile
	s.LoadNumberSpellersFromDir(resourceDir)
	return s
}

// OpenEnglishSynthesizerFromDictPath loads resources from the directory of english_synth.dict.
func OpenEnglishSynthesizerFromDictPath(dictPath string) *EnglishSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenEnglishSynthesizerFromDir(filepath.Dir(dictPath))
}

// loadMergedManual concatenates ManualSynthesizer files (Java dual removal streams).
func loadMergedManual(paths ...string) *synthesis.ManualSynthesizer {
	var buf bytes.Buffer
	any := false
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil || len(data) == 0 {
			continue
		}
		any = true
		buf.Write(data)
		if data[len(data)-1] != '\n' {
			buf.WriteByte('\n')
		}
	}
	if !any {
		return nil
	}
	m, err := synthesis.NewManualSynthesizer(&buf)
	if err != nil {
		return nil
	}
	return m
}
