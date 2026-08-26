package ro

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// roSynthResourceRel is the official LT resource/ro directory layout (Java classpath /ro/...).
const roSynthResourceRel = "inspiration/languagetool/languagetool-language-modules/ro/src/main/resources/org/languagetool/resource/ro"

var (
	instanceOnce sync.Once
	instance     *RomanianSynthesizer
)

// Instance returns RomanianSynthesizer.INSTANCE loaded from official LT resources.
// Nil if romanian_synth.dict cannot be opened (fail-closed).
func Instance() *RomanianSynthesizer {
	instanceOnce.Do(func() {
		if p := DiscoverRomanianSynthDict(); p != "" {
			instance = OpenRomanianSynthesizerFromDictPath(p)
			bindINSTANCE(instance)
		}
	})
	return instance
}

// DiscoverRomanianSynthDict finds romanian_synth.dict (Java RESOURCE_FILENAME /ro/romanian_synth.dict).
// Order: LANG_RO_SYNTH_DICT / LANG_ROMANIAN_SYNTH_DICT, sibling of LANG_ROMANIAN_DICT,
// inspiration module path (dict ships in language-module resources; no third_party jar needed).
func DiscoverRomanianSynthDict() string {
	for _, env := range []string{"LANG_RO_SYNTH_DICT", "LANG_ROMANIAN_SYNTH_DICT"} {
		if p := os.Getenv(env); p != "" {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
	}
	// Sibling of POS dict when only LANG_ROMANIAN_DICT is set.
	if p := os.Getenv("LANG_ROMANIAN_DICT"); p != "" {
		cand := filepath.Join(filepath.Dir(p), "romanian_synth.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
	}
	relPaths := []string{
		filepath.Join(roSynthResourceRel, "romanian_synth.dict"),
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		for _, rel := range relPaths {
			p := filepath.Join(dir, rel)
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// OpenRomanianSynthesizerFromDir loads Java RomanianSynthesizer resources from a
// resource/ro directory containing romanian_synth.dict (+ romanian_tags.txt, added/removed).
// Returns nil if the binary synth dict cannot be opened (fail-closed).
func OpenRomanianSynthesizerFromDir(resourceDir string) *RomanianSynthesizer {
	if resourceDir == "" {
		return nil
	}
	base := synthesis.OpenBaseSynthesizerFromDir("ro", resourceDir)
	if base == nil {
		return nil
	}
	// Preserve Java RESOURCE_FILENAME / TAGS_FILE_NAME for diagnostics (OpenBase sets basename).
	base.ResourceFileName = ResourceFilename
	base.TagFileName = TagsFileName
	// Java TAGS_FILE_NAME is romanian_tags.txt — prefer exact name over glob ambiguity.
	tagsPath := filepath.Join(resourceDir, "romanian_tags.txt")
	if tags, err := synthesis.LoadPossibleTagsFile(tagsPath); err == nil && len(tags) > 0 {
		base.PossibleTags = tags
	}
	// Java BaseSynthesizer also loads /ro/added.txt + /ro/removed.txt from the module tree.
	// Dict ships in inspiration; keep supplement for env/dict-path overrides that only point at the dict.
	supplementManualFromInspiration(base, resourceDir)
	return &RomanianSynthesizer{BaseSynthesizer: base}
}

// OpenRomanianSynthesizerFromDictPath loads resources from the directory of romanian_synth.dict.
func OpenRomanianSynthesizerFromDictPath(dictPath string) *RomanianSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenRomanianSynthesizerFromDir(filepath.Dir(dictPath))
}

// supplementManualFromInspiration loads added/removed from inspiration module when missing
// beside the binary dict (Java classpath merges jar + language-module resources).
func supplementManualFromInspiration(base *synthesis.BaseSynthesizer, resourceDir string) {
	if base == nil {
		return
	}
	insp := findInspirationROResourceDir()
	dirs := []string{}
	if insp != "" && insp != resourceDir {
		dirs = append(dirs, insp)
	}
	if base.Manual == nil {
		for _, d := range dirs {
			if m, err := synthesis.LoadManualSynthesizerFile(filepath.Join(d, "added.txt")); err == nil && m != nil {
				base.Manual = m
				break
			}
		}
	}
	if base.Removal == nil {
		for _, d := range dirs {
			// Java: removed.txt then do-not-synthesize.txt; OpenBase merges when present in resourceDir.
			if m, err := synthesis.LoadManualSynthesizerFile(filepath.Join(d, "removed.txt")); err == nil && m != nil {
				base.Removal = m
				break
			}
		}
	}
}

func findInspirationROResourceDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		cand := filepath.Join(dir, roSynthResourceRel)
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
