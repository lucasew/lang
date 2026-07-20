package synthesis

import (
	"os"
	"path/filepath"
	"strings"
)

// OpenBaseSynthesizerFromDir loads a language synthesizer from a resource directory
// containing *_synth.dict, optional *_tags.txt, added.txt, removed.txt,
// do-not-synthesize.txt. Returns nil if no binary synth dict opens (fail-closed).
func OpenBaseSynthesizerFromDir(langShort, resourceDir string) *BaseSynthesizer {
	if resourceDir == "" {
		return nil
	}
	dictPath := findSynthDictInDir(resourceDir)
	if dictPath == "" {
		return nil
	}
	lookup, err := OpenMorfologikSynthLookup(dictPath)
	if err != nil || lookup == nil {
		return nil
	}
	manual, _ := LoadManualSynthesizerFile(filepath.Join(resourceDir, "added.txt"))
	removal := loadMergedManualFiles(
		filepath.Join(resourceDir, "removed.txt"),
		filepath.Join(resourceDir, "do-not-synthesize.txt"),
	)
	base := NewBaseSynthesizer(langShort, manual)
	base.ResourceFileName = filepath.Base(dictPath)
	// Java sor path is language-specific (e.g. /en/en.sor, /de/de.sor).
	base.SorFileName = "/" + langShort + "/" + langShort + ".sor"
	base.Lookup = lookup
	base.Removal = removal
	base.LoadNumberSpellersFromDir(resourceDir)
	if tags := findTagsFileInDir(resourceDir); tags != "" {
		if list, err := LoadPossibleTagsFile(tags); err == nil && len(list) > 0 {
			base.PossibleTags = list
		}
	}
	return base
}

// OpenBaseSynthesizerFromDictPath loads resources from the directory of a *_synth.dict.
func OpenBaseSynthesizerFromDictPath(langShort, dictPath string) *BaseSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenBaseSynthesizerFromDir(langShort, filepath.Dir(dictPath))
}

func findSynthDictInDir(dir string) string {
	matches, _ := filepath.Glob(filepath.Join(dir, "*_synth.dict"))
	if len(matches) == 0 {
		matches, _ = filepath.Glob(filepath.Join(dir, "*synth.dict"))
	}
	if len(matches) == 0 {
		return ""
	}
	best := matches[0]
	for _, m := range matches[1:] {
		if m < best {
			best = m
		}
	}
	return best
}

func findTagsFileInDir(dir string) string {
	matches, _ := filepath.Glob(filepath.Join(dir, "*_tags.txt"))
	if len(matches) == 0 {
		matches, _ = filepath.Glob(filepath.Join(dir, "*tags.txt"))
	}
	for _, m := range matches {
		base := strings.ToLower(filepath.Base(m))
		if base == "tagset.txt" {
			continue
		}
		return m
	}
	return ""
}

func loadMergedManualFiles(paths ...string) *ManualSynthesizer {
	var parts []byte
	any := false
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil || len(data) == 0 {
			continue
		}
		any = true
		parts = append(parts, data...)
		if data[len(data)-1] != '\n' {
			parts = append(parts, '\n')
		}
	}
	if !any {
		return nil
	}
	m, err := NewManualSynthesizer(strings.NewReader(string(parts)))
	if err != nil {
		return nil
	}
	return m
}
