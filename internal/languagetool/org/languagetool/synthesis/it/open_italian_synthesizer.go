package it

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// itSynthResourceRel is the official LT resource/it directory layout (Java classpath /it/...).
const itSynthResourceRel = "inspiration/languagetool/languagetool-language-modules/it/src/main/resources/org/languagetool/resource/it"

var (
	instanceOnce sync.Once
	instance     *ItalianSynthesizer
)

// Instance returns ItalianSynthesizer.INSTANCE loaded from official LT resources.
// Nil if italian_synth.dict cannot be opened (fail-closed).
func Instance() *ItalianSynthesizer {
	instanceOnce.Do(func() {
		if p := DiscoverItalianSynthDict(); p != "" {
			instance = OpenItalianSynthesizerFromDictPath(p)
			bindINSTANCE(instance)
		}
	})
	return instance
}

// DiscoverItalianSynthDict finds italian_synth.dict (Java RESOURCE_FILENAME /it/italian_synth.dict).
// Order: LANG_IT_SYNTH_DICT / LANG_ITALIAN_SYNTH_DICT, sibling of LANG_ITALIAN_DICT,
// inspiration module path (dict ships in language-module resources; no third_party jar needed).
func DiscoverItalianSynthDict() string {
	for _, env := range []string{"LANG_IT_SYNTH_DICT", "LANG_ITALIAN_SYNTH_DICT"} {
		if p := os.Getenv(env); p != "" {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
	}
	// Sibling of POS dict when only LANG_ITALIAN_DICT is set.
	if p := os.Getenv("LANG_ITALIAN_DICT"); p != "" {
		cand := filepath.Join(filepath.Dir(p), "italian_synth.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
	}
	relPaths := []string{
		filepath.Join(itSynthResourceRel, "italian_synth.dict"),
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

// OpenItalianSynthesizerFromDir loads Java ItalianSynthesizer resources from a
// resource/it directory containing italian_synth.dict (+ italian_tags.txt, added/removed).
// Returns nil if the binary synth dict cannot be opened (fail-closed).
func OpenItalianSynthesizerFromDir(resourceDir string) *ItalianSynthesizer {
	if resourceDir == "" {
		return nil
	}
	base := synthesis.OpenBaseSynthesizerFromDir("it", resourceDir)
	if base == nil {
		return nil
	}
	// Preserve Java RESOURCE_FILENAME / TAGS_FILE_NAME for diagnostics (OpenBase sets basename).
	base.ResourceFileName = ResourceFilename
	base.TagFileName = TagsFileName
	// Java TAGS_FILE_NAME is italian_tags.txt — prefer exact name over glob ambiguity.
	tagsPath := filepath.Join(resourceDir, "italian_tags.txt")
	if tags, err := synthesis.LoadPossibleTagsFile(tagsPath); err == nil && len(tags) > 0 {
		base.PossibleTags = tags
	}
	// Java BaseSynthesizer also loads /it/added.txt + /it/removed.txt from the module tree.
	// Dict ships in inspiration; keep supplement for env/dict-path overrides that only point at the dict.
	supplementManualFromInspiration(base, resourceDir)
	return &ItalianSynthesizer{BaseSynthesizer: base}
}

// OpenItalianSynthesizerFromDictPath loads resources from the directory of italian_synth.dict.
func OpenItalianSynthesizerFromDictPath(dictPath string) *ItalianSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenItalianSynthesizerFromDir(filepath.Dir(dictPath))
}

// supplementManualFromInspiration loads added/removed from inspiration module when missing
// beside the binary dict (Java classpath merges jar + language-module resources).
func supplementManualFromInspiration(base *synthesis.BaseSynthesizer, resourceDir string) {
	if base == nil {
		return
	}
	insp := findInspirationITResourceDir()
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

func findInspirationITResourceDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		cand := filepath.Join(dir, itSynthResourceRel)
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
